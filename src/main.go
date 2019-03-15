package main

import (
	"errors"
	"fmt"
	as "github.com/aerospike/aerospike-client-go"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

func main() {
	var m mainStruct
	os.Exit(m.main())
}

type configStruct struct {
	host    *string
	user    *string
	pass    *string
	nodes   []*string
	command *string
	timeout *time.Duration
}

type mainStruct struct {
	config configStruct
	client *as.Client
}

func (m *mainStruct) main() int {
	if len(os.Args) == 1 || os.Args[1] == "--help" {
		fmt.Printf("Usage: %s -h [HOSTNAME[:IP]] [-u USERNAME -p PASSWORD] [-n NODE_IP,NODE_IP,...] [-t TIMEOUT_MS] COMMAND\n", os.Args[0])
		fmt.Printf("Example: Query and seed from only 'localhost':\n\t%s -n 127.0.0.1 'services'\n", os.Args[0])
		fmt.Printf("Example: Query all nodes for 'services', connect to localhost:\n\t%s 'services'\n", os.Args[0])
		fmt.Printf("Example: As above, timeout 300 seconds\n\t%s -t 300000 'services'\n", os.Args[0])
		fmt.Printf("Example: Connect to 10.0.0.6:3000, use user/pass and get output only from 2 nodes in the cluster\n\t%s -h 10.0.0.6:3000 -u superman -p crypton -n 10.0.0.5,10.0.0.6 -t 300000 'services'\n", os.Args[0])
		return 1
	}
	err := m.parseCommandLine()
	if err != nil {
		log.Fatalf("Error parsing command line parameters: %s", err)
	}
	err = m.connect()
	if err != nil {
		log.Fatalf("Error connecting to cluster: %s", err)
	}
	defer m.client.Close()
	err = m.info()
	if err != nil {
		log.Fatalf("Error running info(): %s", err)
	}
	return 0
}

func (m *mainStruct) parseCommandLine() error {
	for i := 1; i < len(os.Args); i++ {
		if os.Args[i] == "-h" {
			i++
			if i == len(os.Args) {
				return makeError("Usage error: -h is missing host argument")
			}
			m.config.host = &os.Args[i]
		} else if os.Args[i] == "-t" {
			i++
			if i == len(os.Args) {
				return makeError("Usage error: -t is missing timeout seconds argument")
			}
			timeout, err := strconv.Atoi(os.Args[i])
			if err != nil {
				return makeError("Timeout could not be parsed: %s", err)
			}
			to := time.Duration(timeout) * time.Millisecond
			m.config.timeout = &to
		} else if os.Args[i] == "-u" {
			i++
			if i == len(os.Args) {
				return makeError("Usage error: -u is missing user argument")
			}
			m.config.user = &os.Args[i]
		} else if os.Args[i] == "-p" {
			i++
			if i == len(os.Args) {
				return makeError("Usage error: -p is missing password argument")
			}
			m.config.pass = &os.Args[i]
		} else if os.Args[i] == "-n" {
			i++
			if i == len(os.Args) {
				return makeError("Usage error: -n is missing node list argument (comma-separated list of node IPs)")
			}
			nodeList := strings.Split(os.Args[i], ",")
			for j := range nodeList {
				m.config.nodes = append(m.config.nodes, &nodeList[j])
			}
		} else if strings.HasPrefix(os.Args[i], "-") {
			return makeError("Usage error: parameter not recognised: %s", os.Args[i])
		} else {
			if m.config.command != nil {
				return makeError("Command parsing error: already got '%s', now found '%s'", *m.config.command, os.Args[i])
			}
			if i != len(os.Args)-1 {
				return makeError("Command parsing error: found command '%s' and there are trailing arguments after", os.Args[i])
			}
			m.config.command = &os.Args[i]
		}
	}

	if m.config.command == nil {
		return makeError("Command must be specified")
	}
	return nil
}

func (m *mainStruct) connect() error {
	var err error
	var host string
	var port int
	host = "127.0.0.1"
	port = 3000
	if m.config.host != nil {
		hostport := strings.Split(*m.config.host, ":")
		host = hostport[0]
		if len(hostport) > 1 {
			port, err = strconv.Atoi(hostport[1])
			if err != nil {
				return makeError("Error parsing port number: %s", err)
			}
		}
	}
	policy := as.NewClientPolicy()
	if m.config.user != nil {
		policy.User = *m.config.user
		policy.Password = *m.config.pass
	}
	if m.config.timeout != nil {
		policy.Timeout = *m.config.timeout
		policy.IdleTimeout = *m.config.timeout
	}
	nTime := time.Now()
	m.client, err = as.NewClientWithPolicy(policy, host, port)
	connectTime := time.Since(nTime)
	if err != nil {
		return makeError("Error connecting: %s", err)
	}
	fmt.Printf("Connected in %0.3fs\n\n", connectTime.Seconds())
	return nil
}

func (m *mainStruct) info() error {
	nodeList := m.client.GetNodes()
	var wg sync.WaitGroup
	for _, node := range nodeList {
		runNode := false
		if m.config.nodes != nil {
			nodeIp := node.GetHost().Name
			for _, confNode := range m.config.nodes {
				if nodeIp == *confNode {
					runNode = true
					break
				}
			}
		}
		if m.config.nodes == nil || runNode == true {
			wg.Add(1)
			go func() {
				defer wg.Done()
				nTime := time.Now()
				out, err := node.RequestInfo(*m.config.command)
				infoTime := time.Since(nTime)
				nout := fmt.Sprintf("NODE %s (%s:%d) RESPONDED IN %0.3fs:\n", node.GetName(), node.GetHost().Name, node.GetHost().Port, infoTime.Seconds())
				if err != nil {
					nout = fmt.Sprintf("%sError running info command: %s\n", nout, err)
				} else {
					for _, v := range out {
						nout = fmt.Sprintf("%s%s", nout, fmt.Sprintln(v))
					}
				}
				nout = fmt.Sprintf("%s\n", nout)
				fmt.Print(nout)
			}()
		}
	}
	wg.Wait()
	return nil
}

func makeError(name string, args ...interface{}) error {
	return errors.New(fmt.Sprintf(name, args...))
}

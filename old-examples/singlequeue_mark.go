package main

import (
	"flag"
	"fmt"
	"github.com/elico/go-nfqueue"
	"os"
	"os/signal"
	"sync/atomic"
)

var marksMax uint64
var logpkt bool
var logmark bool
var queueNum int

/*
func print_packets(qid uint16, pkt *nfqueue.Packet) {
	fmt.Println(pkt)
	pkt.Accept()
}
*/

func main() {
	flag.BoolVar(&logpkt, "log-packet", false, "Log the packet to stdout (works with log-mark option only)")
	flag.BoolVar(&logmark, "log-mark", false, "Log the mark selection to stdout")

	flag.Uint64Var(&marksMax, "high-mark", uint64(3), "The number of the highest queue number")
	flag.IntVar(&queueNum, "queue-num", 0, "The NFQUEQUE number")

	flag.Parse()
	var (
		q = nfqueue.NewNFQueue(uint16(queueNum))
	)
	defer q.Destroy()

	fmt.Println("The queue is active, add an iptables rule to use it, for example: ")
	fmt.Println("\tiptables -t mangle -I PREROUTING 1 [-i eth0] -m conntrack --ctstate NEW -j NFQUEUE --queue-num", queueNum)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)
	packets := q.Process()
	counter := uint64(1)

LOOP:
	for {
		select {
		case pkt := <-packets:
			val := (atomic.AddUint64(&counter, 1) % marksMax) + 1
			if val == uint64(0) {
				val++
			}
			if logmark {
				if logpkt {
					fmt.Println("The selected Mark =>", val, "For packet =>", pkt)
				} else {
					fmt.Println("The selected Mark =>", val)
				}
			}
			pkt.RepeatMark(uint32(val))
		case <-sig:
			break LOOP
		}

	}
	fmt.Println("Exiting, remember to remove the iptables rule :")
	fmt.Println("\tiptables -D INPUT -m conntrack --ctstate NEW -j NFQUEUE --queue-num", queueNum)
}

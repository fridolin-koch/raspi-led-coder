package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/host"
)

var (
	flagPin              = flag.String("pin", "7", "Output Pin")
	flagCodeWordDuration = flag.Uint("duration", 50, "Code Word duration in milliseconds")
	flagPayload          = flag.Uint("payload", 100, "Payload")
	flagRepeat           = flag.Uint("repeat", 0, "Usage repeat after milliseconds")
)

func main() {
	// parse flags
	flag.Parse()
	// Load all the drivers:
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}
	payload := uint8(*flagPayload)
	fmt.Printf("Payload: %d (%08b)\n", payload, payload)
	// convert duration
	duration := time.Duration(*flagCodeWordDuration) * time.Millisecond
	repeatDuration := time.Duration(*flagRepeat) * time.Millisecond
	// get pin
	pin := gpioreg.ByName(*flagPin)
	if pin == nil {
		log.Fatalf("Pin %s is not available", *flagPin)
	}
	// get bits
	bits := fmt.Sprintf("%08b", payload)

	fmt.Printf("Payload Coded: %s\n", encode(bits))

	// clean shutdown
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sigc
		if err := powerOff(pin); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()

	for {
		log.Println("Start")
		// send preamble
		if err := send("101", duration*2, pin); err != nil {
			log.Fatal(err)
		}
		// send message
		if err := send(bits, duration, pin); err != nil {
			log.Fatal(err)
		}
		// send trailer
		if err := send("101", duration*2, pin); err != nil {
			log.Fatal(err)
		}
		log.Println("End")
		if repeatDuration == 0 {
			break
		}
		time.Sleep(repeatDuration)
	}
	// shutdown
	if err := powerOff(pin); err != nil {
		log.Fatal(err)
	}
}

func send(message string, pause time.Duration, pin gpio.PinIO) error {
	message = encode(message)
	for _, bit := range message {
		l := gpio.Low
		if bit == '1' {
			l = gpio.High
		}
		if err := pin.Out(l); err != nil {
			return err
		}
		log.Printf("Sending %s", string(bit))
		time.Sleep(pause)
	}
	return nil
}

func powerOff(pin gpio.PinIO) error {
	if err := pin.Out(gpio.Low); err != nil {
		return err
	}
	return nil
}

// taken from https://github.com/mahdavipanah/hamcode/blob/master/hamcode.go#L152
// Returns encoded data code as an Hamming code
func encode(data string) (hcode string) {
	bits := make([]int, 0)

	for pos, i := 0, 0; i < len(data); pos++ {
		if isPerfectSquare(pos + 1) {
			bits = append(bits, 0)
		} else {
			bits = append(bits, cToB(data[i]))
			i++
		}
	}

	for pos, _ := range bits {
		if isPerfectSquare(pos + 1) {
			p := 0
			for i, _ := range bits {
				// Checks if the bit should be calculated
				if i+1 != pos+1 && ((i+1)&(pos+1) != 0) {
					p ^= bits[i]
				}
			}

			bits[pos] = p
		}
	}

	for _, bit := range bits {
		hcode += string(bit + 48)
	}

	return

}

// Checks if a number is perfect square
func isPerfectSquare(n int) bool {
	return n == (n & -n)
}

// Converts char{0,1} to int{0,1}
func cToB(b byte) int {
	if b == '1' {
		return 1
	}
	return 0
}

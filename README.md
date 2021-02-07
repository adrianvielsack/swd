This project implements all elements needed to flash a STM32 from within a go application without the need for external components.


## Flashing a stm32: 
```go
import (
	"github.com/adrianvielsack/swd"
	"github.com/adrianvielsack/swd/lowlevel"
)

// Frequency determines the approx delay between clock signal level changes
iface, err := lowlevel.NewSwdRaspberrygpio(18, 22, 10000000)
if err != nil {
    log.Fatalln("Could not initialize GPIOs")
    return nil
}
controller := swd.NewSTM32(swd.NewSWD(iface))

err := swd.FlashSTM32(controller, data)
if err != nil {
    log.Fatal(err)
}


```
# fanatec-pedal-emulator

Small Arduino sketch and golang HID proxy

Allows a set of USB pedals to be proxied to a Fanatec wheelbase and therefore
be used on console if the wheelbase supports it

# build/run proxy

    go run proxy.go

# build emulator

1. install the HID-Project library


# microchip PIC

CSL Elite LC uses the following PIC18F26J53


# fanatec RJ12 pinouts on pedal control board

Looking from the top of the socket

|Socket / Pin | 1    | 2      | 3    | 4  | 5  | 6   |
|-------------|------|--------|------|----|----|-----|
|Gas          | 3.3v | Signal | NC   | NC | NC | GND |
|Brake        | 3.3v | S-     | S+   | NC | NC | GND |
|Clutch       | 3.3v | Signal | NC   | NC | NC | GND |
|WheelBase    | GND? | GND?   | GND? | RX | TX | Vcc |

Note RX / TX is from PoV of the control board




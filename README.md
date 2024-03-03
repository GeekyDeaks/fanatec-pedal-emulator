# fanatec-pedal-emulator

Small golang HID to UART proxy

Allows a set of USB pedals to be proxied to a Fanatec wheelbase and therefore
be used on console if the wheelbase supports it

# setup

The Fanatec CSL Elite pedal control module uses a simple UART protocol to
communicate with the wheelbase.  A 5v tolerant USB to RS232 TTL like the CP2102
can be used to interface with the pedal RJ12 port on the wheelbase.

# build/run proxy

    go run proxy.go [COMPORT]

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




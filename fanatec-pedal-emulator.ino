#include "HID-Project.h"

// this is the raw HID report from the proxy
typedef struct {
  //uint8_t id;
  uint8_t throttle;
  uint8_t brake;
  uint8_t clutch;
  uint8_t handbrake;
} USBReport;

USBReport report;
uint8_t * report_ptr = (uint8_t *) &report;

uint8_t rawhidData[64];

void setup() {

  Serial1.begin(230400);
  Serial.begin(115200);

  // Set the RawHID OUT report array.
  // Feature reports are also (parallel) possible, see the other example for this.
  RawHID.begin(rawhidData, sizeof(rawhidData));

}

uint32_t pkt = 0;

void loop() {

  // Check if there is new data from the RawHID device
  auto bytesAvailable = RawHID.available();
  if (bytesAvailable)
  {
    for(uint8_t i = 0; i < bytesAvailable; i++) {
      // copy the data to the struct if we are within range
      if(i < sizeof(USBReport)) {
        report_ptr[i] = RawHID.read();
      } else {
         RawHID.read();
      }
    }
    Serial.print(pkt++);
    Serial.print(": ");
    //Serial.print(report.id);
    //Serial.print(" ");
    Serial.print(report.throttle);
    Serial.print(" ");
    Serial.print(report.brake);
    Serial.print(" ");
    Serial.println(report.clutch);
  }

  if (Serial1.available() > 0)  {
    // wheelbase sent a request
    int fb;

    do {
       fb = Serial1.read(); 
    } while(Serial1.available() > 0);

    byte out = 0x80;
    // what did it ask for?
    switch(fb) {
      case 0x80: // brake?
        out = report.clutch | 0x80;
        break;
      case 0xA0:
        // throttle 0 - 4096
       out = report.throttle | 0x80;
        break;
      case 0xC0:
        // 
        out = report.brake | 0x80;
        
        break;
      case 0xE0:
        // handbrake?
        out = report.handbrake | 0x80;
        break;
      
    }
    Serial1.write(out);
  }
}

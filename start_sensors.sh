#!/bin/bash

xterm -e go run src/distributed/sensors/sensor.go -name=boiler_pressure_out -min=15 -max=15.5 -step=0.05 -freq=1 &
xterm -e go run src/distributed/sensors/sensor.go -name=turbine_pressure_out -min=0.9 -max=1.3 -step=0.05 -freq=1 &
xterm -e go run src/distributed/sensors/sensor.go -name=condensor_pressure_out -min=0.001 -max=0.002 -step=0.0001 -freq=1 &
xterm -e go run src/distributed/sensors/sensor.go -name=boiler_temp_out -min=590 -max=615 -step=1 -freq=1 &
xterm -e go run src/distributed/sensors/sensor.go -name=turbine_temp_out -min=100 -max=105 -step=1 -freq=1 &
xterm -e go run src/distributed/sensors/sensor.go -name=condensor_temp_out -min=80  -max=98  -step=1 -freq=1 &
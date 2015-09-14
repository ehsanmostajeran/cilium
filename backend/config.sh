#!/bin/bash

BRIDGE="lxc-br0"
TUNNEL="vx0"
HOSTPORT="host0"
OFVERSION="OpenFlow13,OpenFlow12,OpenFlow11,OpenFlow10"

LOGICAL_ROUTER=192.0.2.1
LOGICAL_ROUTER_MAC=dd:dd:dd:dd:dd:dd

# Pipeline
TBL_PRE=0
TBL_MAIN=1
TBL_POLICY=2

# Default broadcast domain
BD_MAIN=1

#####################################################################
# Register usage
REG_SGRP=reg0
REG_SGRP_OF="NXM_NX_REG0[]"
REG_DGRP=reg1
REG_DGRP_OF="NXM_NX_REG1[]"
REG_BD=reg2
REG_BD_OF="NXM_NX_REG2[]"
REG_NS=reg3
REG_NS_OF="NXM_NX_REG3[]"
REG_PORT=reg4
REG_PORT_OF="NXM_NX_REG4[]"


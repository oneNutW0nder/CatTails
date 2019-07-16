# CatTails

## Overview  
This project is a redteam implant that leverages raw sockets to 
send/recieve callbacks from a C2 infrastructure.  
  
The callbacks and communication happen over UDP port 53. You will be able to send/execute commands on the remote host via a C2 server.  

### Features

-  **custom commands***
    - Ex. If you want to flush `iptables` CatTails will provide an
      abstraction for you to do this. Instead of sending the shell commands
      necessary you will be able to run a CatTails command like `drop-rules`. 
- **Command feedback/output***
    - CatTails will send you the output of a command (if there is any)  
      and let you know if your command completed successfully.  

(*) Work in progress  
(x) Completed

*More coming soon!*

### v.Alpha Roadmap
- [x] RAW sockets that allow bypass of host-based firewalls 
- [x] Bots dynamically determine gateway MAC, IP, etc. during install
- [x] IP/UDP Header lengths and checksums are properly computed
- [x] Packets are properly filtered 
- [ ] Payloads are parsed and handled based on C2 or bot needs
- [ ] Bots will send HELLO messages on a time interval, providing the C2 with client info
- [ ] C2 stores/tracks client information from bot HELLOs
- [ ] C2 simple CLI to interface with bots (show info, # of hosts, send commands, etc)
- [ ] Bots return output of commands executed up to MTU (fragmentation not supported yet)


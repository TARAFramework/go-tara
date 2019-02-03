# go-tara

[![Join the chat at https://gitter.im/TARAFramework/go-tara](https://badges.gitter.im/TARAFramework/go-tara.svg)](https://gitter.im/TARAFramework/go-tara?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

[![Join our ZenDesk at https://app.zenhub.com/workspaces/tara-dev-5c3e1c26ca7fa74504b9153c/boards](https://img.shields.io/badge/Shipping_faster_with-ZenHub-5e60ba.svg?style=flat-square)](https://app.zenhub.com/workspaces/tara-dev-5c3e1c26ca7fa74504b9153c/boards)

Welcome to the implementation of TARA Framework in go language!  
This reference implementation is heavy WIP.  
Please use it with due considerations.

## TARA CLI (PoC)

### Installation

1. Get the go-tara repo
  
   ``` sh
   go get github.com/TARAFramework/go-tara
   ```

2. Enter the go-tara repo in your $GOPATH/src
  
   ``` sh
   cd $GOPATH/src/github.com/TARAFramework/go-tara/
   ```

3. Run make for the executable
  
   ``` sh
   make go-tara
   ```

### Usage

1. Navigate to `go-tara` repo as discussed above
  
2. Open 3 terminals & execute the following commands one for each:
  
   ``` sh
    ./go-tara -nodeid 1
   ```
  
   ``` sh
    ./go-tara -nodeid 2
   ```
  
   ``` sh
    ./go-tara -nodeid 3
   ```

### Inputs

1. The buffer will open on the terminal console for your inputs.
  
2. *If inputs end with a '?' character*, the client request is relayed to `Cluster 1` encoded by its key ID 100.
  
3. *If inputs **do not** end with a '?' character*, the client request is relayed to `Cluster 2` encoded by its key ID 101.

<!-- //TODO: ### Customizability -->

## Other stuffs @ TARA

1. [Tara Paper](https://github.com/TARAFramework/tara-paper/blob/master/paper.pdf)  - Working Draft by [@cattitude](https://github.com/cattitude). Presents the complete concepts, design approaches for a scalable RAFT Architecture that can handle more complex incoming client requests.
2. [TaraScope](https://taraframework.github.io/tarascope/) - Visual representation of a behavior of triplets.
3. [TARA.tla](https://github.com/TARAFramework/tara.tla) - The TLA+ Specification for RAFT along with tests.

## Credits (and Thanks!)

The go-tara proof of concept is based on [@lni](https://github.com/lni/)'s dragonboat multigroup implementation.  
Link for the same is here: [https://github.com/lni/dragonboat](https://github.com/lni/dragonboat) 💚

# TriBFT Consensus Framework

A highly scalable Byzantine fault-tolerant consensus framework that provides a three-layer architecture for vehicular network blockchain and vehicular reputation evaluation model for large-scale vehicular networking.

## ⚠️ Notice

This repository contains the **core module code** for the paper:

> **TriBFT-IoV: A Trustworthy Consensus System Based on Adaptive Three-Layer Architecture for Large-Scale Internet of Vehicles**

The code is provided for **academic review purposes only**. 

**This project is developed based on [BlockEmulator](https://github.com/HuangLab-SYSU/block-emulator). A complete runtime environment requires integration with BlockEmulator.**

## 🏗️ Architecture

TriBFT implements a three-layer virtual architecture:

```
┌─────────────────────────────────────────┐
│           Global Shard Layer            │
│      (Cross-shard Coordination)         │
├─────────────────────────────────────────┤
│          City Cluster Layer             │
│       (Regional Aggregation)            │
├─────────────────────────────────────────┤
│         Regional Shard Layer            │
│        (Actual Execution)               │
└─────────────────────────────────────────┘
```

## ✨ Key Features

- **HotStuff Consensus Engine**: O(n) communication complexity with pipelined block processing
- **VRM (Verifiable Reputation Mechanism)**: Dual-layer reputation model (global + local)
- **Dynamic Sharding**: Adaptive shard splitting and merging based on network load
- **Three-layer Architecture**: Hierarchical consensus for improved scalability
- **Inter-layer Delay Simulation**: Realistic multi-server deployment simulation

## 📂 Project Structure

```
TriBFT-Consensus-Framework/
├── consensus_shard/tribft/     # TriBFT core implementation
│   ├── tribft_node.go          # Main node (three-layer integration)
│   ├── hotstuff.go             # HotStuff consensus engine
│   ├── hotstuff_log.go         # Block log and commit rules
│   ├── city_aggregator.go      # City aggregator
│   └── global_store.go         # Global reputation storage
├── reputation/vrm/             # VRM reputation mechanism
│   ├── reputation_calculator.go
│   ├── local_reputation_manager.go
│   └── global_reputation_store.go
├── chain/                      # Blockchain core (blocks, tx pool)
├── message/                    # Message type definitions
├── networks/                   # P2P network communication
├── params/                     # Global configuration
└── supervisor/                 # Supervisor node (experiment data collection)
```

## 📋 Requirements

- Go 1.19+
- BlockEmulator integration (for full runtime)

## 🚀 Build

```bash
go build -o tribft main.go
```

## 📖 References

- [BlockEmulator](https://github.com/HuangLab-SYSU/block-emulator) - The base blockchain emulator
- [HotStuff: BFT Consensus with Linearity and Responsiveness](https://arxiv.org/abs/1803.05069)

## 📄 License

This project is licensed under the BSD 2-Clause License - see the [LICENSE](LICENSE) file for details.

## 📧 Contact

For questions about this project, please contact the authors.

---

**Note**: This is a preview version. Full implementation will be released after project completion.

# 🦙 **kubectllama** - AI-powered Kubernetes CLI

Welcome to **kubectllama**! 🐾 The AI-powered CLI tool that takes your Kubernetes management to the next level by allowing you to run `kubectl` commands through **natural language**. 🎉 Say goodbye to memorizing complex `kubectl` commands and let **kubectllama** handle it for you! 🤖✨

---

## 🌟 Features

- 🗣️ **Natural Language Processing**: Simply type commands like "Get all pods in the default namespace" and let **kubectllama** do the magic.
- ⚡ **Fast & Efficient**: Get complex `kubectl` commands with minimal effort and increased productivity.
- 🔒 **Safe & Secure**: Your AI assistant lives locally on your machine, ensuring your commands are processed securely.
- 💬 **Confirmation Step**: the cli doesn't execute any command, **kubectllama** will only display the suggested command so to prevent unwanted actions.

---

## 🛠️ Installation

`kubectllama` can be installed either by downloading a pre-built executable from GitHub Releases or by cloning the repository and building from source. Below are instructions for both methods.

### Prerequisites
- **Ollama**: You need Ollama installed and running locally (default URL: `http://localhost:11434`). Download it from [Ollama's website](https://ollama.com/). 
  - **Model**: By default, `kubectllama` uses the `mistral` model because it offers the best tradeoff between speed and precision. However, you can use any Ollama model by specifying it with the `--model` flag (e.g., `--model llama3`). Pull your chosen model:
    ```bash
    ollama pull mistral  # Default model
    ```
    Or, for a different model:
    ```bash
    ollama pull <model-name>
    ```
- **Go**: Required only for building from source (version 1.21+).

### Method 1: Download Pre-Built Executable
Pre-built binaries are available for Linux, macOS, and Windows from the [GitHub Releases page](https://github.com/your-username/kllama/releases). Since the repository is public, no authentication is needed.

#### Linux
```bash
curl -L -o kubectllama \
 https://github.com/jazzshu/kubectllama/releases/latest/download/kubectllama-linux-amd64
chmod +x kubectllama
sudo mv kubectllama /usr/local/bin/
```

#### macOS
```bash
curl -L -o kubectllama \
     https://github.com/jazzshu/kubectllama/releases/latest/download/kubectllama-macos-amd64
chmod +x kubectllama
sudo mv kubectllama /usr/local/bin/
```

#### Windows

1. Download kubectllama-windows-amd64.exe from the latest release.
2. Move it to a directory in your PATH (e.g., C:\Windows\System32) using File Explorer or:
```bash
move kubectllama-windows-amd64.exe C:\Windows\System32\kubectllama.exe
```

#### Verify Installation
```bash
kubectllama --help
```

### Method 2: Clone and Build from Source
If you prefer to build kubectllama yourself or want to modify the code:

1. **Clone the repository**:
```bash
git clone https://github.com/jazzshu/kubectllama.git
cd kubectllama
```

2. **Build the binary**:
```bash
go build -o kubectllama .
```

3. **Install it**:
```bash
sudo mv kubectllama /usr/local/bin
chmod +x /usr/local/bin/kubectllama
```
 - For Windows, move it to a PATH directory:
  ```bash
    move kubectllama.exe C:\Windows\System32\
  ```
#### Verify Installation
```bash
kubectllama --help
```

## 🚀 Usage
After installation, run ```kubectllama``` with a natural language request:

```bash
kubectllama get pods running in the test namespace
```

Output (using default ```mistral``` model):
```bash
--> kubectl get pods -n test
```

To use a different model, specify it with the ```--model``` flag:
```bash
kubectllama --model llama3 get pods running in the test namespace
```

If Ollama is running a different host from the default one, you can specify it with the ```--url``` flag:
```bash
kubectllama --url http://my-ollama-custom-url:8080 get pods running in test namespace
```


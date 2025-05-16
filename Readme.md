
# Site Analyzer

Site Analyzer is a simple Go web application that accepts a URL and performs the following analyses:
- Determines the HTML version
- Extracts and displays the page title
- Counts the number of headings (h1–h6)
- Identifies the presence of a login form
- Extracts links and categorizes them as internal/external and accessible/inaccessible



## Prerequisites

- Docker installed on your system
- Internet access to pull Go and Alpine images and analyze external sites

## How to Build and Run with Docker

### 1. Clone the Repository

- git clone https://github.com/manjulaD/site-analyzer.git
- cd site-analyzer


### 2. Build the Docker Image

```bash
docker build -t site-analyzer .
```

### 3. Run the Container

```bash
docker run -p 8080:8080 site-analyzer
```

Now, access the application in your browser at:

```
http://localhost:8080
```



<?php require_once('lightspeed/version.php'); ?>
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Documentation | Lightspeed</title>
    <link rel="icon" type="image/svg+xml" href="data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 48 48'><circle cx='24' cy='24' r='22' fill='%23deb841'/><circle cx='24' cy='24' r='10' fill='white'/><circle cx='24' cy='24' r='7' fill='%23deb841'/></svg>">
    <link rel="stylesheet" href="../assets/css/style.css">
    <link rel="stylesheet" href="../assets/css/docs.css">
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700;800&family=JetBrains+Mono:wght@400;500&display=swap" rel="stylesheet">
</head>
<body class="docs-page">
    <header class="site-header">
        <div class="container header-inner">
            <a href="../" class="logo">
                <svg class="logo-icon" width="28" height="28" viewBox="0 0 48 48">
                    <circle cx="24" cy="24" r="22" fill="var(--accent)"/>
                    <circle cx="24" cy="24" r="10" fill="white"/>
                    <circle cx="24" cy="24" r="7" fill="var(--accent)"/>
                </svg>
                <span class="logo-text">Lightspeed</span>
            </a>
            <nav class="main-nav">
                <ul class="nav-menu">
                    <li><a href="../#features">Features</a></li>
                    <li><a href="../#how-it-works">Get Started</a></li>
                    <li><a href="./" class="active">Documentation</a></li>
                    <li><a href="../#showcase">Showcase</a></li>
                </ul>
                <a href="../#how-it-works" class="btn btn-primary">Develop Now</a>
            </nav>
            <button class="mobile-menu-toggle" aria-label="Toggle menu">
                <span></span>
                <span></span>
                <span></span>
            </button>
        </div>
    </header>

    <div class="docs-layout">
        <aside class="docs-sidebar">
            <div class="docs-version">v<?php echo lightspeed_version(); ?></div>
            <nav class="docs-nav">
                <div class="docs-nav-section">
                    <h3>Getting Started</h3>
                    <ul>
                        <li><a href="#introduction" class="active">Introduction</a></li>
                        <li><a href="#installation">Installation</a></li>
                        <li><a href="#quick-start">Quick Start</a></li>
                        <li><a href="#requirements">Requirements</a></li>
                    </ul>
                </div>
                <div class="docs-nav-section">
                    <h3>CLI Commands</h3>
                    <ul>
                        <li><a href="#cmd-init">init</a></li>
                        <li><a href="#cmd-start">start</a></li>
                        <li><a href="#cmd-stop">stop</a></li>
                        <li><a href="#cmd-build">build</a></li>
                        <li><a href="#cmd-publish">publish</a></li>
                        <li><a href="#cmd-deploy">deploy</a></li>
                    </ul>
                </div>
                <div class="docs-nav-section">
                    <h3>Configuration</h3>
                    <ul>
                        <li><a href="#site-properties">site.properties</a></li>
                        <li><a href="#config-options">Configuration Options</a></li>
                        <li><a href="#custom-domains">Custom Domains</a></li>
                        <li><a href="#environment-vars">Environment Variables</a></li>
                    </ul>
                </div>
                <div class="docs-nav-section">
                    <h3>Project Structure</h3>
                    <ul>
                        <li><a href="#directory-layout">Directory Layout</a></li>
                        <li><a href="#assets">Assets &amp; Static Files</a></li>
                        <li><a href="#includes">PHP Includes</a></li>
                        <li><a href="#routing">Routing &amp; Clean URLs</a></li>
                    </ul>
                </div>
                <div class="docs-nav-section">
                    <h3>PHP Library</h3>
                    <ul>
                        <li><a href="#php-library">Using the Library</a></li>
                        <li><a href="#library-functions">Available Functions</a></li>
                        <li><a href="#ide-support">IDE Support</a></li>
                    </ul>
                </div>
                <div class="docs-nav-section">
                    <h3>Deployment</h3>
                    <ul>
                        <li><a href="#deploy-workflow">Deployment Workflow</a></li>
                        <li><a href="#docker-images">Docker Images</a></li>
                        <li><a href="#server-image">Server Image</a></li>
                        <li><a href="#ssl-cdn">SSL &amp; CDN</a></li>
                    </ul>
                </div>
                <div class="docs-nav-section">
                    <h3>Advanced</h3>
                    <ul>
                        <li><a href="#versioning">Versioning</a></li>
                        <li><a href="#libraries">Custom Libraries</a></li>
                        <li><a href="#troubleshooting">Troubleshooting</a></li>
                    </ul>
                </div>
            </nav>
        </aside>

        <main class="docs-content">
            <div class="docs-container">
                <!-- Introduction -->
                <section id="introduction" class="docs-section">
                    <h1>Lightspeed Documentation</h1>
                    <p class="lead">Lightspeed is a lightweight, rapid development tool for building and deploying PHP websites. Get from zero to production in minutes with our streamlined CLI workflow.</p>

                    <div class="feature-highlight">
                        <div class="feature-highlight-item">
                            <h4>Simple Framework</h4>
                            <p>No complex configurations. Just HTML, CSS, PHP, and JavaScript.</p>
                        </div>
                        <div class="feature-highlight-item">
                            <h4>Instant Deployment</h4>
                            <p>Deploy your site with a single command to our global infrastructure.</p>
                        </div>
                        <div class="feature-highlight-item">
                            <h4>Built-in Dev Server</h4>
                            <p>Local development server with hot reload and clean URLs.</p>
                        </div>
                    </div>
                </section>

                <!-- Installation -->
                <section id="installation" class="docs-section">
                    <h2>Installation</h2>
                    <p>Install the Lightspeed CLI to get started. Choose the installation method for your platform.</p>

                    <h3>macOS / Linux</h3>
                    <p>Install with a single command using curl:</p>
                    <div class="code-block">
                        <div class="code-header">
                            <span>Terminal</span>
                            <button class="copy-btn" aria-label="Copy code">Copy</button>
                        </div>
                        <pre><code>curl -sfL https://raw.githubusercontent.com/abrayall/lightspeed/refs/heads/main/install.sh | sh -</code></pre>
                    </div>

                    <h3>Windows (PowerShell)</h3>
                    <p>On Windows, use PowerShell to install:</p>
                    <div class="code-block">
                        <div class="code-header">
                            <span>PowerShell</span>
                            <button class="copy-btn" aria-label="Copy code">Copy</button>
                        </div>
                        <pre><code>irm https://raw.githubusercontent.com/abrayall/lightspeed/refs/heads/main/install.ps1 | iex</code></pre>
                    </div>

                    <h3>From Source</h3>
                    <p>Alternatively, you can build from source:</p>
                    <div class="code-block">
                        <div class="code-header">
                            <span>Terminal</span>
                            <button class="copy-btn" aria-label="Copy code">Copy</button>
                        </div>
                        <pre><code>git clone https://github.com/abrayall/lightspeed.git
cd lightspeed
./install.sh</code></pre>
                    </div>

                    <h3>Verify Installation</h3>
                    <p>After installation, verify that Lightspeed is installed correctly:</p>
                    <div class="code-block">
                        <div class="code-header">
                            <span>Terminal</span>
                            <button class="copy-btn" aria-label="Copy code">Copy</button>
                        </div>
                        <pre><code>lightspeed --version</code></pre>
                    </div>
                </section>

                <!-- Quick Start -->
                <section id="quick-start" class="docs-section">
                    <h2>Quick Start</h2>
                    <p>Get your first Lightspeed site up and running in under 5 minutes.</p>

                    <h3>1. Create a New Project</h3>
                    <p>Create a directory for your project and initialize it:</p>
                    <div class="code-block">
                        <div class="code-header">
                            <span>Terminal</span>
                            <button class="copy-btn" aria-label="Copy code">Copy</button>
                        </div>
                        <pre><code>mkdir my-awesome-site && cd my-awesome-site
lightspeed init</code></pre>
                    </div>
                    <p>This creates the basic project structure with a starter <code>index.php</code>, CSS files, and configuration.</p>

                    <h3>2. Start the Development Server</h3>
                    <p>Launch the local development server:</p>
                    <div class="code-block">
                        <div class="code-header">
                            <span>Terminal</span>
                            <button class="copy-btn" aria-label="Copy code">Copy</button>
                        </div>
                        <pre><code>lightspeed start</code></pre>
                    </div>
                    <p>Your site is now running at <code>http://localhost:9000</code>. Changes to your files are reflected immediately.</p>

                    <h3>3. Build Your Site</h3>
                    <p>Edit your PHP, HTML, CSS, and JavaScript files. The development server supports:</p>
                    <ul>
                        <li><strong>Clean URLs</strong> - Access <code>/about</code> instead of <code>/about.php</code></li>
                        <li><strong>Hot reload</strong> - Changes appear instantly in the browser</li>
                        <li><strong>PHP includes</strong> - Organize your code with includes</li>
                    </ul>

                    <h3>4. Deploy to Production</h3>
                    <p>When you're ready to go live, deploy with a single command:</p>
                    <div class="code-block">
                        <div class="code-header">
                            <span>Terminal</span>
                            <button class="copy-btn" aria-label="Copy code">Copy</button>
                        </div>
                        <pre><code>lightspeed deploy</code></pre>
                    </div>
                    <p>Your site will be live at <code>https://my-awesome-site.lightspeed.ee</code></p>
                </section>

                <!-- Requirements -->
                <section id="requirements" class="docs-section">
                    <h2>Requirements</h2>
                    <p>Before using Lightspeed, ensure you have the following installed:</p>

                    <div class="requirements-grid">
                        <div class="requirement-card">
                            <h4>Docker</h4>
                            <p>Required for the development server and building production images. <a href="https://www.docker.com/get-started" target="_blank">Install Docker</a></p>
                        </div>
                        <div class="requirement-card">
                            <h4>DigitalOcean Account</h4>
                            <p>Required for deployment to Lightspeed hosting. <a href="https://www.digitalocean.com" target="_blank">Create Account</a></p>
                        </div>
                    </div>

                    <h3>Supported Platforms</h3>
                    <ul>
                        <li>macOS (Intel and Apple Silicon)</li>
                        <li>Linux (x64 and ARM64)</li>
                        <li>Windows 10/11 (with WSL2 recommended)</li>
                    </ul>
                </section>

                <!-- CLI Commands -->
                <section id="cmd-init" class="docs-section">
                    <h2>CLI Commands</h2>

                    <h3>lightspeed init</h3>
                    <p>Initialize a new Lightspeed project with the basic directory structure and starter files.</p>

                    <div class="code-block">
                        <div class="code-header">
                            <span>Terminal</span>
                            <button class="copy-btn" aria-label="Copy code">Copy</button>
                        </div>
                        <pre><code>lightspeed init
lightspeed init --name mysite
lightspeed init --name mysite --domain example.com
lightspeed init -d example.com -d www.example.com</code></pre>
                    </div>

                    <h4>Options</h4>
                    <table class="options-table">
                        <thead>
                            <tr>
                                <th>Option</th>
                                <th>Description</th>
                                <th>Default</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr>
                                <td><code>-n, --name</code></td>
                                <td>Site name</td>
                                <td>Directory name</td>
                            </tr>
                            <tr>
                                <td><code>-d, --domain</code></td>
                                <td>Domain(s) for the site. Can be specified multiple times.</td>
                                <td>name.com</td>
                            </tr>
                        </tbody>
                    </table>

                    <h4>Files Created</h4>
                    <ul>
                        <li><code>site.properties</code> - Site configuration</li>
                        <li><code>index.php</code> - Hello World starter page</li>
                        <li><code>assets/css/style.css</code> - Basic stylesheet</li>
                        <li><code>assets/js/</code> - JavaScript directory</li>
                        <li><code>includes/</code> - PHP includes directory</li>
                        <li><code>.idea/</code> - PhpStorm project configuration</li>
                        <li><code>.gitignore</code> - Git ignore file</li>
                    </ul>

                    <div class="info-box">
                        <strong>Note:</strong> Running <code>init</code> again is safe - it only creates files that don't exist and updates the PhpStorm configuration.
                    </div>
                </section>

                <section id="cmd-start" class="docs-section">
                    <h3>lightspeed start</h3>
                    <p>Start a PHP development server using Docker. The server mounts your current directory and serves it with hot reload capabilities.</p>

                    <div class="code-block">
                        <div class="code-header">
                            <span>Terminal</span>
                            <button class="copy-btn" aria-label="Copy code">Copy</button>
                        </div>
                        <pre><code>lightspeed start
lightspeed start --port 8080
lightspeed start --image custom-server</code></pre>
                    </div>

                    <h4>Options</h4>
                    <table class="options-table">
                        <thead>
                            <tr>
                                <th>Option</th>
                                <th>Description</th>
                                <th>Default</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr>
                                <td><code>-p, --port</code></td>
                                <td>Port to expose</td>
                                <td>Auto-detect in 9000 range</td>
                            </tr>
                            <tr>
                                <td><code>-i, --image</code></td>
                                <td>Docker image to use</td>
                                <td>lightspeed-server</td>
                            </tr>
                        </tbody>
                    </table>

                    <h4>Features</h4>
                    <ul>
                        <li><strong>Clean URLs</strong> - Access <code>/about</code> instead of <code>/about.php</code></li>
                        <li><strong>Automatic PHP library loading</strong> - Libraries from <code>~/.lightspeed/library/</code></li>
                        <li><strong>Hot reload</strong> - Changes are reflected immediately</li>
                    </ul>
                </section>

                <section id="cmd-stop" class="docs-section">
                    <h3>lightspeed stop</h3>
                    <p>Stop the running development server.</p>

                    <div class="code-block">
                        <div class="code-header">
                            <span>Terminal</span>
                            <button class="copy-btn" aria-label="Copy code">Copy</button>
                        </div>
                        <pre><code>lightspeed stop</code></pre>
                    </div>
                </section>

                <section id="cmd-build" class="docs-section">
                    <h3>lightspeed build</h3>
                    <p>Build a Docker container for production deployment. The image is optimized for the <code>linux/amd64</code> platform.</p>

                    <div class="code-block">
                        <div class="code-header">
                            <span>Terminal</span>
                            <button class="copy-btn" aria-label="Copy code">Copy</button>
                        </div>
                        <pre><code>lightspeed build
lightspeed build --tag 1.0.0
lightspeed build --image custom-base-image</code></pre>
                    </div>

                    <h4>Options</h4>
                    <table class="options-table">
                        <thead>
                            <tr>
                                <th>Option</th>
                                <th>Description</th>
                                <th>Default</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr>
                                <td><code>-t, --tag</code></td>
                                <td>Version tag</td>
                                <td>Git version or 'latest'</td>
                            </tr>
                            <tr>
                                <td><code>-i, --image</code></td>
                                <td>Base Docker image</td>
                                <td>lightspeed-server</td>
                            </tr>
                        </tbody>
                    </table>
                </section>

                <section id="cmd-publish" class="docs-section">
                    <h3>lightspeed publish</h3>
                    <p>Build and push the Docker image to the Lightspeed registry. This pushes both the versioned tag and a <code>latest</code> tag.</p>

                    <div class="code-block">
                        <div class="code-header">
                            <span>Terminal</span>
                            <button class="copy-btn" aria-label="Copy code">Copy</button>
                        </div>
                        <pre><code>lightspeed publish
lightspeed publish --tag 1.0.0
lightspeed publish --name my-site</code></pre>
                    </div>

                    <h4>Options</h4>
                    <table class="options-table">
                        <thead>
                            <tr>
                                <th>Option</th>
                                <th>Description</th>
                                <th>Default</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr>
                                <td><code>-t, --tag</code></td>
                                <td>Version tag</td>
                                <td>Git version or 'latest'</td>
                            </tr>
                            <tr>
                                <td><code>-n, --name</code></td>
                                <td>Site name</td>
                                <td>From site.properties or directory name</td>
                            </tr>
                        </tbody>
                    </table>
                </section>

                <section id="cmd-deploy" class="docs-section">
                    <h3>lightspeed deploy</h3>
                    <p>Build, push, and deploy your site to production. If the app doesn't exist, it will be created automatically.</p>

                    <div class="code-block">
                        <div class="code-header">
                            <span>Terminal</span>
                            <button class="copy-btn" aria-label="Copy code">Copy</button>
                        </div>
                        <pre><code>lightspeed deploy
lightspeed deploy --name my-production-site</code></pre>
                    </div>

                    <h4>Options</h4>
                    <table class="options-table">
                        <thead>
                            <tr>
                                <th>Option</th>
                                <th>Description</th>
                                <th>Default</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr>
                                <td><code>-n, --name</code></td>
                                <td>Site name</td>
                                <td>Project directory name</td>
                            </tr>
                        </tbody>
                    </table>

                    <p>After deployment, your site will be accessible at:</p>
                    <ul>
                        <li><code>https://[name].lightspeed.ee</code> - Automatically configured subdomain</li>
                    </ul>
                </section>

                <!-- Configuration -->
                <section id="site-properties" class="docs-section">
                    <h2>Configuration</h2>

                    <h3>site.properties</h3>
                    <p>The <code>site.properties</code> file in your project root configures your site. Here's a complete example:</p>

                    <div class="code-block">
                        <div class="code-header">
                            <span>site.properties</span>
                            <button class="copy-btn" aria-label="Copy code">Copy</button>
                        </div>
                        <pre><code># Site name (used for [name].lightspeed.ee)
name=mysite

# Custom domains
domain=example.com
domains=www.example.com,app.example.com

# Base image (pin to specific version)
image=0.5.4

# PHP libraries (for include path)
libraries=lightspeed</code></pre>
                    </div>
                </section>

                <section id="config-options" class="docs-section">
                    <h3>Configuration Options</h3>

                    <table class="options-table">
                        <thead>
                            <tr>
                                <th>Property</th>
                                <th>Description</th>
                                <th>Default</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr>
                                <td><code>name</code></td>
                                <td>Site name, used for lightspeed.ee subdomain</td>
                                <td>Directory name</td>
                            </tr>
                            <tr>
                                <td><code>domain</code></td>
                                <td>Single custom domain</td>
                                <td>-</td>
                            </tr>
                            <tr>
                                <td><code>domains</code></td>
                                <td>Comma-separated list of custom domains</td>
                                <td>-</td>
                            </tr>
                            <tr>
                                <td><code>image</code></td>
                                <td>Base Docker image version</td>
                                <td>CLI version</td>
                            </tr>
                            <tr>
                                <td><code>libraries</code></td>
                                <td>Comma-separated PHP library paths</td>
                                <td>-</td>
                            </tr>
                        </tbody>
                    </table>

                    <h4>Image Property Examples</h4>
                    <div class="code-block">
                        <div class="code-header">
                            <span>site.properties</span>
                            <button class="copy-btn" aria-label="Copy code">Copy</button>
                        </div>
                        <pre><code># Use specific version
image=0.5.4

# Use latest
image=latest

# Use custom image
image=ghcr.io/myorg/myimage:latest</code></pre>
                    </div>

                    <h4>Libraries Property Examples</h4>
                    <div class="code-block">
                        <div class="code-header">
                            <span>site.properties</span>
                            <button class="copy-btn" aria-label="Copy code">Copy</button>
                        </div>
                        <pre><code># Use lightspeed library (matches CLI version)
libraries=lightspeed

# Use specific lightspeed version
libraries=lightspeed:v0.5.0

# Multiple libraries
libraries=lightspeed,/path/to/custom/lib</code></pre>
                    </div>
                </section>

                <section id="custom-domains" class="docs-section">
                    <h3>Custom Domains</h3>
                    <p>You can configure custom domains for your site in <code>site.properties</code>:</p>

                    <div class="code-block">
                        <div class="code-header">
                            <span>site.properties</span>
                            <button class="copy-btn" aria-label="Copy code">Copy</button>
                        </div>
                        <pre><code># Single domain
domain=example.com

# Multiple domains
domains=www.example.com,app.example.com,api.example.com</code></pre>
                    </div>

                    <h4>DNS Configuration</h4>
                    <p>Point your custom domain to Lightspeed by creating a CNAME record:</p>
                    <table class="options-table">
                        <thead>
                            <tr>
                                <th>Type</th>
                                <th>Name</th>
                                <th>Value</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr>
                                <td>CNAME</td>
                                <td>www</td>
                                <td>your-site.lightspeed.ee</td>
                            </tr>
                            <tr>
                                <td>CNAME</td>
                                <td>@</td>
                                <td>your-site.lightspeed.ee</td>
                            </tr>
                        </tbody>
                    </table>

                    <div class="info-box">
                        <strong>SSL Certificates:</strong> SSL certificates are automatically provisioned for all custom domains using Let's Encrypt.
                    </div>
                </section>

                <section id="environment-vars" class="docs-section">
                    <h3>Environment Variables</h3>
                    <p>You can pass environment variables to your application for configuration:</p>

                    <div class="code-block">
                        <div class="code-header">
                            <span>PHP</span>
                            <button class="copy-btn" aria-label="Copy code">Copy</button>
                        </div>
                        <pre><code>&lt;?php
// Access environment variables
$apiKey = getenv('API_KEY');
$debugMode = getenv('DEBUG') === 'true';
?&gt;</code></pre>
                    </div>

                    <p>Common environment variables available in your Lightspeed application:</p>
                    <table class="options-table">
                        <thead>
                            <tr>
                                <th>Variable</th>
                                <th>Description</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr>
                                <td><code>LIGHTSPEED_ENV</code></td>
                                <td>Current environment (development, production)</td>
                            </tr>
                            <tr>
                                <td><code>LIGHTSPEED_VERSION</code></td>
                                <td>Lightspeed version</td>
                            </tr>
                        </tbody>
                    </table>
                </section>

                <!-- Project Structure -->
                <section id="directory-layout" class="docs-section">
                    <h2>Project Structure</h2>

                    <h3>Directory Layout</h3>
                    <p>A typical Lightspeed project has the following structure:</p>

                    <div class="code-block">
                        <div class="code-header">
                            <span>Project Structure</span>
                        </div>
                        <pre><code>mysite/
├── site.properties     # Site configuration
├── index.php           # Main entry point
├── about.php           # Additional pages
├── contact.php
├── assets/
│   ├── css/
│   │   └── style.css   # Stylesheets
│   ├── js/
│   │   └── main.js     # JavaScript files
│   └── images/         # Image assets
├── includes/           # PHP includes
│   ├── header.php
│   ├── footer.php
│   └── functions.php
├── .idea/              # PhpStorm configuration
│   └── php.xml         # PHP include paths
├── .gitignore          # Git ignore file
└── Dockerfile          # Generated on build</code></pre>
                    </div>
                </section>

                <section id="assets" class="docs-section">
                    <h3>Assets &amp; Static Files</h3>
                    <p>Static assets like CSS, JavaScript, and images are served from the <code>assets/</code> directory:</p>

                    <div class="code-block">
                        <div class="code-header">
                            <span>HTML</span>
                            <button class="copy-btn" aria-label="Copy code">Copy</button>
                        </div>
                        <pre><code>&lt;!-- CSS --&gt;
&lt;link rel="stylesheet" href="/assets/css/style.css"&gt;

&lt;!-- JavaScript --&gt;
&lt;script src="/assets/js/main.js"&gt;&lt;/script&gt;

&lt;!-- Images --&gt;
&lt;img src="/assets/images/logo.png" alt="Logo"&gt;</code></pre>
                    </div>

                    <h4>Recommended Organization</h4>
                    <ul>
                        <li><code>assets/css/</code> - Stylesheets</li>
                        <li><code>assets/js/</code> - JavaScript files</li>
                        <li><code>assets/images/</code> - Images and icons</li>
                        <li><code>assets/fonts/</code> - Custom fonts</li>
                    </ul>
                </section>

                <section id="includes" class="docs-section">
                    <h3>PHP Includes</h3>
                    <p>The <code>includes/</code> directory is for PHP files that are included by other files. This is perfect for headers, footers, and shared functions.</p>

                    <div class="code-block">
                        <div class="code-header">
                            <span>includes/header.php</span>
                            <button class="copy-btn" aria-label="Copy code">Copy</button>
                        </div>
                        <pre><code>&lt;!DOCTYPE html&gt;
&lt;html lang="en"&gt;
&lt;head&gt;
    &lt;meta charset="UTF-8"&gt;
    &lt;title&gt;&lt;?php echo $pageTitle ?? 'My Site'; ?&gt;&lt;/title&gt;
    &lt;link rel="stylesheet" href="/assets/css/style.css"&gt;
&lt;/head&gt;
&lt;body&gt;
    &lt;header&gt;
        &lt;nav&gt;...&lt;/nav&gt;
    &lt;/header&gt;</code></pre>
                    </div>

                    <div class="code-block">
                        <div class="code-header">
                            <span>index.php</span>
                            <button class="copy-btn" aria-label="Copy code">Copy</button>
                        </div>
                        <pre><code>&lt;?php
$pageTitle = 'Home';
include 'includes/header.php';
?&gt;

&lt;main&gt;
    &lt;h1&gt;Welcome to my site&lt;/h1&gt;
&lt;/main&gt;

&lt;?php include 'includes/footer.php'; ?&gt;</code></pre>
                    </div>
                </section>

                <section id="routing" class="docs-section">
                    <h3>Routing &amp; Clean URLs</h3>
                    <p>Lightspeed automatically provides clean URLs without the <code>.php</code> extension:</p>

                    <table class="options-table">
                        <thead>
                            <tr>
                                <th>File</th>
                                <th>URL</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr>
                                <td><code>index.php</code></td>
                                <td><code>/</code></td>
                            </tr>
                            <tr>
                                <td><code>about.php</code></td>
                                <td><code>/about</code></td>
                            </tr>
                            <tr>
                                <td><code>contact.php</code></td>
                                <td><code>/contact</code></td>
                            </tr>
                            <tr>
                                <td><code>blog/index.php</code></td>
                                <td><code>/blog</code></td>
                            </tr>
                            <tr>
                                <td><code>blog/post.php</code></td>
                                <td><code>/blog/post</code></td>
                            </tr>
                        </tbody>
                    </table>

                    <div class="info-box">
                        <strong>Tip:</strong> You can still access pages with the <code>.php</code> extension if needed, but clean URLs are recommended for SEO and user experience.
                    </div>
                </section>

                <!-- PHP Library -->
                <section id="php-library" class="docs-section">
                    <h2>PHP Library</h2>

                    <h3>Using the Library</h3>
                    <p>Lightspeed includes a PHP library that's automatically available in the server image at <code>/opt/lightspeed/</code>. The PHP include path is configured automatically.</p>

                    <div class="code-block">
                        <div class="code-header">
                            <span>PHP</span>
                            <button class="copy-btn" aria-label="Copy code">Copy</button>
                        </div>
                        <pre><code>&lt;?php
require_once('lightspeed/version.php');

echo lightspeed_version(); // Returns the Lightspeed version
?&gt;</code></pre>
                    </div>

                    <p>The library is downloaded to <code>~/.lightspeed/library/v[version]/</code> on first use.</p>
                </section>

                <section id="library-functions" class="docs-section">
                    <h3>Available Functions</h3>
                    <p>The Lightspeed PHP library provides several utility functions:</p>

                    <table class="options-table">
                        <thead>
                            <tr>
                                <th>Function</th>
                                <th>Description</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr>
                                <td><code>lightspeed_version()</code></td>
                                <td>Returns the current Lightspeed version</td>
                            </tr>
                        </tbody>
                    </table>
                </section>

                <section id="ide-support" class="docs-section">
                    <h3>IDE Support</h3>
                    <p>When you run <code>lightspeed init</code> or any lightspeed command in a project with <code>.idea/</code> and <code>site.properties</code>, the PhpStorm include paths are automatically updated to point to the resolved library locations.</p>

                    <p>This enables:</p>
                    <ul>
                        <li>Code completion for Lightspeed library functions</li>
                        <li>Go-to-definition for library code</li>
                        <li>Error checking for undefined functions</li>
                    </ul>

                    <div class="info-box">
                        <strong>Note:</strong> The <code>.idea/php.xml</code> file is automatically generated and should be committed to version control.
                    </div>
                </section>

                <!-- Deployment -->
                <section id="deploy-workflow" class="docs-section">
                    <h2>Deployment</h2>

                    <h3>Deployment Workflow</h3>
                    <p>The typical deployment workflow consists of three steps:</p>

                    <div class="workflow-steps">
                        <div class="workflow-step">
                            <div class="step-num">1</div>
                            <div class="step-info">
                                <h4>Build</h4>
                                <p>Create a Docker image with your site code</p>
                                <code>lightspeed build</code>
                            </div>
                        </div>
                        <div class="workflow-step">
                            <div class="step-num">2</div>
                            <div class="step-info">
                                <h4>Publish</h4>
                                <p>Push the image to the Lightspeed registry</p>
                                <code>lightspeed publish</code>
                            </div>
                        </div>
                        <div class="workflow-step">
                            <div class="step-num">3</div>
                            <div class="step-info">
                                <h4>Deploy</h4>
                                <p>Deploy to production infrastructure</p>
                                <code>lightspeed deploy</code>
                            </div>
                        </div>
                    </div>

                    <p>Or, use the all-in-one command that performs all three steps:</p>
                    <div class="code-block">
                        <div class="code-header">
                            <span>Terminal</span>
                            <button class="copy-btn" aria-label="Copy code">Copy</button>
                        </div>
                        <pre><code>lightspeed deploy</code></pre>
                    </div>
                </section>

                <section id="docker-images" class="docs-section">
                    <h3>Docker Images</h3>
                    <p>Lightspeed builds production Docker images optimized for PHP hosting. The build process:</p>

                    <ol>
                        <li>Creates a <code>Dockerfile</code> in your project root</li>
                        <li>Copies your site files to <code>/var/www/html/</code></li>
                        <li>Sets appropriate permissions</li>
                        <li>Tags the image with your site name and version</li>
                    </ol>

                    <div class="code-block">
                        <div class="code-header">
                            <span>Generated Dockerfile</span>
                        </div>
                        <pre><code>FROM ghcr.io/abrayall/lightspeed-server:0.6.3
COPY . /var/www/html/
RUN chown -R www-data:www-data /var/www/html</code></pre>
                    </div>
                </section>

                <section id="server-image" class="docs-section">
                    <h3>Server Image</h3>
                    <p>Lightspeed uses a custom server image (<code>ghcr.io/abrayall/lightspeed-server</code>) based on:</p>

                    <ul>
                        <li><strong>PHP 8.2 FPM</strong> - Latest PHP with performance optimizations</li>
                        <li><strong>Nginx</strong> - High-performance web server</li>
                    </ul>

                    <h4>Features</h4>
                    <ul>
                        <li>Clean URLs (no <code>.php</code> extension required)</li>
                        <li>Pre-configured PHP include path for Lightspeed library</li>
                        <li>Optimized for small PHP sites</li>
                        <li>Gzip compression enabled</li>
                        <li>Security headers configured</li>
                    </ul>
                </section>

                <section id="ssl-cdn" class="docs-section">
                    <h3>SSL &amp; CDN</h3>
                    <p>All Lightspeed deployments include:</p>

                    <div class="feature-highlight">
                        <div class="feature-highlight-item">
                            <h4>Free SSL Certificates</h4>
                            <p>Automatic SSL certificates via Let's Encrypt for all domains.</p>
                        </div>
                        <div class="feature-highlight-item">
                            <h4>Global CDN</h4>
                            <p>Content delivery network for fast page loads worldwide.</p>
                        </div>
                        <div class="feature-highlight-item">
                            <h4>99.9% Uptime SLA</h4>
                            <p>Enterprise-grade infrastructure with high availability.</p>
                        </div>
                    </div>
                </section>

                <!-- Advanced -->
                <section id="versioning" class="docs-section">
                    <h2>Advanced</h2>

                    <h3>Versioning</h3>
                    <p>Lightspeed uses git tags for versioning. When you build or publish, the version is automatically determined from your git history:</p>

                    <div class="code-block">
                        <div class="code-header">
                            <span>Terminal</span>
                            <button class="copy-btn" aria-label="Copy code">Copy</button>
                        </div>
                        <pre><code># Tag a release
git tag v1.0.0
git push --tags

# Build with the version
lightspeed build
# Image: registry.lightspeed.ee/mysite:1.0.0</code></pre>
                    </div>

                    <p>If no git tag is found, the version defaults to <code>latest</code>.</p>
                </section>

                <section id="libraries" class="docs-section">
                    <h3>Custom Libraries</h3>
                    <p>You can add custom PHP libraries to your project by specifying them in <code>site.properties</code>:</p>

                    <div class="code-block">
                        <div class="code-header">
                            <span>site.properties</span>
                            <button class="copy-btn" aria-label="Copy code">Copy</button>
                        </div>
                        <pre><code># Multiple libraries
libraries=lightspeed,/path/to/custom/lib,/another/library</code></pre>
                    </div>

                    <p>Libraries are added to the PHP include path, so you can use <code>require_once()</code> with relative paths:</p>

                    <div class="code-block">
                        <div class="code-header">
                            <span>PHP</span>
                            <button class="copy-btn" aria-label="Copy code">Copy</button>
                        </div>
                        <pre><code>&lt;?php
require_once('mylib/helper.php');
require_once('anotherlib/utils.php');
?&gt;</code></pre>
                    </div>
                </section>

                <section id="troubleshooting" class="docs-section">
                    <h3>Troubleshooting</h3>

                    <h4>Docker not running</h4>
                    <p>If you see an error about Docker, make sure Docker Desktop is running:</p>
                    <div class="code-block">
                        <div class="code-header">
                            <span>Terminal</span>
                            <button class="copy-btn" aria-label="Copy code">Copy</button>
                        </div>
                        <pre><code># Check if Docker is running
docker info

# Start Docker Desktop (macOS)
open -a Docker</code></pre>
                    </div>

                    <h4>Port already in use</h4>
                    <p>If port 9000 is already in use, specify a different port:</p>
                    <div class="code-block">
                        <div class="code-header">
                            <span>Terminal</span>
                            <button class="copy-btn" aria-label="Copy code">Copy</button>
                        </div>
                        <pre><code>lightspeed start --port 8080</code></pre>
                    </div>

                    <h4>Permission denied</h4>
                    <p>If you see permission errors during install, you may need to use sudo:</p>
                    <div class="code-block">
                        <div class="code-header">
                            <span>Terminal</span>
                            <button class="copy-btn" aria-label="Copy code">Copy</button>
                        </div>
                        <pre><code>sudo ./install.sh</code></pre>
                    </div>

                    <h4>Getting Help</h4>
                    <p>If you encounter issues not covered here:</p>
                    <ul>
                        <li>Check the <a href="https://github.com/abrayall/lightspeed/issues">GitHub Issues</a> for known problems</li>
                        <li>Run <code>lightspeed --help</code> for command usage</li>
                        <li>Contact support for deployment issues</li>
                    </ul>
                </section>

            </div>
        </main>
    </div>

    <footer class="site-footer">
        <div class="container">
            <div class="footer-content">
                <div class="footer-brand">
                    <a href="../" class="logo">
                        <svg class="logo-icon" width="28" height="28" viewBox="0 0 48 48">
                            <circle cx="24" cy="24" r="22" fill="var(--accent)"/>
                            <circle cx="24" cy="24" r="10" fill="white"/>
                            <circle cx="24" cy="24" r="7" fill="var(--accent)"/>
                        </svg>
                        <span class="logo-text">Lightspeed</span>
                    </a>
                    <p>Build, deploy, and host websites at <span style="color: var(--accent); font-weight: 700;">Lightspeed</span>.</p>
                </div>
                <div class="footer-links">
                    <div class="footer-col">
                        <h4>Product</h4>
                        <ul>
                            <li><a href="../#features">Features</a></li>
                            <li><a href="../#showcase">Showcase</a></li>
                        </ul>
                    </div>
                    <div class="footer-col">
                        <h4>Documentation</h4>
                        <ul>
                            <li><a href="./">Getting Started</a></li>
                            <li><a href="./#cmd-init">CLI Commands</a></li>
                            <li><a href="./#site-properties">Configuration</a></li>
                        </ul>
                    </div>
                </div>
            </div>
            <div class="footer-bottom">
                <p>&copy; <?php echo date('Y'); ?> <span style="color: var(--accent);">Lightspeed</span>. All rights reserved.</p>
                <p>Built and hosted with <span style="color: var(--accent);">Lightspeed</span> v<?php echo lightspeed_version(); ?></p>
            </div>
        </div>
    </footer>

    <script src="../assets/js/main.js"></script>
    <script>
        // Docs navigation highlighting
        document.addEventListener('DOMContentLoaded', function() {
            const sections = document.querySelectorAll('.docs-section');
            const navLinks = document.querySelectorAll('.docs-nav a');

            function updateActiveLink() {
                let current = '';
                sections.forEach(section => {
                    const sectionTop = section.offsetTop - 100;
                    if (window.scrollY >= sectionTop) {
                        current = section.getAttribute('id');
                    }
                });

                navLinks.forEach(link => {
                    link.classList.remove('active');
                    if (link.getAttribute('href') === '#' + current) {
                        link.classList.add('active');
                    }
                });
            }

            window.addEventListener('scroll', updateActiveLink);
            updateActiveLink();
        });
    </script>
</body>
</html>

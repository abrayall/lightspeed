<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Lightspeed | Fast Website Hosting & Development Platform</title>
    <link rel="icon" type="image/svg+xml" href="data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 48 48'><circle cx='24' cy='24' r='22' fill='%23deb841'/><circle cx='24' cy='24' r='10' fill='white'/><circle cx='24' cy='24' r='7' fill='%23deb841'/></svg>">
    <link rel="stylesheet" href="assets/css/style.css">
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700;800&family=JetBrains+Mono:wght@400;500&display=swap" rel="stylesheet">
</head>
<body>
    <header class="site-header">
        <div class="container header-inner">
            <a href="#" class="logo">
                <svg class="logo-icon" width="28" height="28" viewBox="0 0 48 48">
                    <circle cx="24" cy="24" r="22" fill="var(--accent)"/>
                    <circle cx="24" cy="24" r="10" fill="white"/>
                    <circle cx="24" cy="24" r="7" fill="var(--accent)"/>
                </svg>
                <span class="logo-text">Lightspeed</span>
            </a>
            <nav class="main-nav">
                <ul class="nav-menu">
                    <li><a href="#features">Features</a></li>
                    <li><a href="#how-it-works">How It Works</a></li>
                    <li><a href="#pricing">Pricing</a></li>
                    <li><a href="#showcase">Showcase</a></li>
                </ul>
                <a href="#get-started" class="btn btn-primary">Get Started</a>
            </nav>
            <button class="mobile-menu-toggle" aria-label="Toggle menu">
                <span></span>
                <span></span>
                <span></span>
            </button>
        </div>
    </header>

    <main>
        <!-- Hero Section -->
        <section class="hero">
            <div class="hero-bg">
                <div class="slide" data-name="Earth at Night"></div>
                <div class="slide" data-name="Mountain"></div>
            </div>
            <div class="container">
                <div class="hero-content">
                    <h1>Develop. Test. Deploy.<br><span class="highlight">At Lightspeed.</span></h1>
                    <p>The fastest way to build and host websites. Simple development framework, instant deployment, and reliable hosting for just $29/month.</p>
                    <div class="hero-cta">
                        <a href="#get-started" class="btn btn-primary btn-lg">Start Building</a>
                        <a href="#how-it-works" class="btn btn-outline btn-lg">Learn More</a>
                    </div>
                    <div class="hero-terminal">
                        <div class="terminal-header">
                            <span class="terminal-dot red"></span>
                            <span class="terminal-dot yellow"></span>
                            <span class="terminal-dot green"></span>
                        </div>
                        <div class="terminal-body">
                            <code><span class="prompt">$</span> lightspeed init mysite</code>
                            <code class="output">Creating new Lightspeed project...</code>
                            <code class="output success">Done! Your site is ready.</code>
                            <code><span class="prompt">$</span> lightspeed start</code>
                            <code class="output success">Server running at http://localhost:9000</code>
                        </div>
                    </div>
                </div>
            </div>
        </section>

        <!-- Features Section -->
        <section id="features" class="features-section">
            <div class="container">
                <div class="section-header">
                    <span class="section-tag">Why Lightspeed?</span>
                    <h2>Everything you need to build amazing websites</h2>
                    <p>A complete platform for developers who want to move fast without sacrificing quality.</p>
                </div>
                <div class="features-grid">
                    <div class="feature-card">
                        <div class="feature-icon">
                            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                <polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/>
                            </svg>
                        </div>
                        <h3>Lightning Fast</h3>
                        <p>From zero to deployed in minutes. Our streamlined workflow eliminates friction so you can focus on building.</p>
                    </div>
                    <div class="feature-card">
                        <div class="feature-icon">
                            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                <path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z"/>
                            </svg>
                        </div>
                        <h3>Simple Framework</h3>
                        <p>No complex configurations. Just HTML, CSS, PHP, and JavaScript. Build the way you want to build.</p>
                    </div>
                    <div class="feature-card">
                        <div class="feature-icon">
                            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                <rect x="2" y="3" width="20" height="14" rx="2" ry="2"/>
                                <line x1="8" y1="21" x2="16" y2="21"/>
                                <line x1="12" y1="17" x2="12" y2="21"/>
                            </svg>
                        </div>
                        <h3>Local Development</h3>
                        <p>Built-in development server with hot reload. See your changes instantly as you code.</p>
                    </div>
                    <div class="feature-card">
                        <div class="feature-icon">
                            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                <path d="M17.5 19H9a7 7 0 1 1 6.71-9h1.79a4.5 4.5 0 1 1 0 9Z"/>
                            </svg>
                        </div>
                        <h3>One-Click Deploy</h3>
                        <p>Push your site live with a single command. No FTP, no complicated pipelines.</p>
                    </div>
                    <div class="feature-card">
                        <div class="feature-icon">
                            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
                            </svg>
                        </div>
                        <h3>Reliable Hosting</h3>
                        <p>99.9% uptime guarantee with automatic SSL certificates and global CDN included.</p>
                    </div>
                    <div class="feature-card">
                        <div class="feature-icon">
                            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                                <circle cx="12" cy="12" r="10"/>
                                <circle cx="12" cy="12" r="4"/>
                                <path d="M12 2a10 10 0 0 1 8.66 5H12"/>
                                <path d="M2.34 7a10 10 0 0 1 4.66-5l5 8.66"/>
                                <path d="M7 21.66a10 10 0 0 1-4.66-14.66l5 8.66"/>
                            </svg>
                        </div>
                        <h3>AI Ready</h3>
                        <p>Built to work seamlessly with AI tools. Let your favorite AI assistant help you build and customize your site.</p>
                    </div>
                </div>
            </div>
        </section>

        <!-- How It Works Section -->
        <section id="how-it-works" class="how-section">
            <div class="container">
                <div class="section-header">
                    <span class="section-tag">Get Started</span>
                    <h2>Up and running in 3 simple steps</h2>
                    <p>From installation to live site in just minutes.</p>
                </div>

                <div class="steps-container">
                    <div class="step">
                        <div class="step-number">1</div>
                        <div class="step-content">
                            <h3>Install Lightspeed</h3>
                            <p>Install the Lightspeed CLI with a single command using npm or your favorite package manager.</p>
                            <div class="code-block">
                                <div class="code-header">
                                    <span>Terminal</span>
                                    <button class="copy-btn" aria-label="Copy code">Copy</button>
                                </div>
                                <pre><code><span class="prompt">$</span> npm install -g @aspect/lightspeed</code></pre>
                            </div>
                        </div>
                    </div>

                    <div class="step">
                        <div class="step-number">2</div>
                        <div class="step-content">
                            <h3>Create Your Project</h3>
                            <p>Initialize a new project with the built-in scaffolding. Choose from templates or start fresh.</p>
                            <div class="code-block">
                                <div class="code-header">
                                    <span>Terminal</span>
                                    <button class="copy-btn" aria-label="Copy code">Copy</button>
                                </div>
                                <pre><code><span class="prompt">$</span> lightspeed init my-awesome-site
<span class="output">Creating project structure...</span>
<span class="output">Installing dependencies...</span>
<span class="output success">Project created successfully!</span>

<span class="prompt">$</span> cd my-awesome-site
<span class="prompt">$</span> lightspeed start
<span class="output success">Development server running at http://localhost:9000</span></code></pre>
                            </div>
                        </div>
                    </div>

                    <div class="step">
                        <div class="step-number">3</div>
                        <div class="step-content">
                            <h3>Deploy & Go Live</h3>
                            <p>When you're ready, deploy your site with one command. It's live in seconds.</p>
                            <div class="code-block">
                                <div class="code-header">
                                    <span>Terminal</span>
                                    <button class="copy-btn" aria-label="Copy code">Copy</button>
                                </div>
                                <pre><code><span class="prompt">$</span> lightspeed deploy
<span class="output">Building for production...</span>
<span class="output">Uploading to Lightspeed Cloud...</span>
<span class="output success">Deployed! Your site is live at:</span>
<span class="output url">https://my-awesome-site.lightspeed.dev</span></code></pre>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </section>

        <!-- Pricing Section -->
        <section id="pricing" class="pricing-section">
            <div class="container">
                <div class="section-header">
                    <span class="section-tag">Simple Pricing</span>
                    <h2>One plan. Everything included.</h2>
                    <p>No hidden fees, no surprises. Just straightforward pricing.</p>
                </div>

                <div class="pricing-card">
                    <div class="pricing-header">
                        <h3>Pro Hosting</h3>
                        <div class="price">
                            <span class="currency">$</span>
                            <span class="amount">29</span>
                            <span class="period">/month</span>
                        </div>
                        <p>Everything you need to host your site</p>
                    </div>
                    <ul class="pricing-features">
                        <li><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><polyline points="20 6 9 17 4 12"/></svg> Unlimited bandwidth</li>
                        <li><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><polyline points="20 6 9 17 4 12"/></svg> Free SSL certificate</li>
                        <li><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><polyline points="20 6 9 17 4 12"/></svg> Global CDN</li>
                        <li><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><polyline points="20 6 9 17 4 12"/></svg> Custom domain</li>
                        <li><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><polyline points="20 6 9 17 4 12"/></svg> Automatic backups</li>
                        <li><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><polyline points="20 6 9 17 4 12"/></svg> 24/7 support</li>
                        <li><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><polyline points="20 6 9 17 4 12"/></svg> 99.9% uptime SLA</li>
                    </ul>
                    <a href="#get-started" class="btn btn-primary btn-lg btn-block">Start Your Free Trial</a>
                    <p class="pricing-note">14-day free trial. No credit card required.</p>
                </div>
            </div>
        </section>

        <!-- Showcase Section -->
        <section id="showcase" class="showcase-section">
            <div class="container">
                <div class="section-header">
                    <span class="section-tag">Showcase</span>
                    <h2>Built with Lightspeed</h2>
                    <p>See what developers are creating with our platform.</p>
                </div>

                <div class="showcase-grid">
                    <div class="showcase-item">
                        <div class="showcase-image">
                            <div class="showcase-placeholder">
                                <span>JustFlow</span>
                            </div>
                        </div>
                        <div class="showcase-info">
                            <h4>JustFlow Events & Marketing</h4>
                            <p>Marketing agency website</p>
                        </div>
                    </div>
                    <div class="showcase-item">
                        <div class="showcase-image">
                            <div class="showcase-placeholder">
                                <span>TechStart</span>
                            </div>
                        </div>
                        <div class="showcase-info">
                            <h4>TechStart Labs</h4>
                            <p>Startup landing page</p>
                        </div>
                    </div>
                    <div class="showcase-item">
                        <div class="showcase-image">
                            <div class="showcase-placeholder">
                                <span>Portfolio</span>
                            </div>
                        </div>
                        <div class="showcase-info">
                            <h4>Sarah Chen Design</h4>
                            <p>Designer portfolio</p>
                        </div>
                    </div>
                    <div class="showcase-item">
                        <div class="showcase-image">
                            <div class="showcase-placeholder">
                                <span>Blog</span>
                            </div>
                        </div>
                        <div class="showcase-info">
                            <h4>Dev Thoughts</h4>
                            <p>Technical blog</p>
                        </div>
                    </div>
                </div>
            </div>
        </section>

        <!-- CTA Section -->
        <section id="get-started" class="cta-section">
            <div class="container">
                <div class="cta-content">
                    <h2>Ready to build at lightspeed?</h2>
                    <p>Join thousands of developers who ship faster with Lightspeed.</p>
                    <div class="cta-form">
                        <input type="email" placeholder="Enter your email" class="email-input">
                        <button class="btn btn-light">Get Started Free</button>
                    </div>
                    <p class="cta-note">Free to develop locally. Only pay when you're ready to host.</p>
                </div>
            </div>
        </section>
    </main>

    <footer class="site-footer">
        <div class="container">
            <div class="footer-content">
                <div class="footer-brand">
                    <a href="#" class="logo">
                        <svg class="logo-icon" width="28" height="28" viewBox="0 0 48 48">
                            <circle cx="24" cy="24" r="22" fill="var(--accent)"/>
                            <circle cx="24" cy="24" r="10" fill="white"/>
                            <circle cx="24" cy="24" r="7" fill="var(--accent)"/>
                        </svg>
                        <span class="logo-text">Lightspeed</span>
                    </a>
                    <p>Build, deploy, and host websites at <span style="color: var(--accent); font-weight: 700;">lightspeed</span>.</p>
                </div>
                <div class="footer-links">
                    <div class="footer-col">
                        <h4>Product</h4>
                        <ul>
                            <li><a href="#features">Features</a></li>
                            <li><a href="#pricing">Pricing</a></li>
                            <li><a href="#showcase">Showcase</a></li>
                        </ul>
                    </div>
                    <div class="footer-col">
                        <h4>Resources</h4>
                        <ul>
                            <li><a href="#">Documentation</a></li>
                            <li><a href="#">Tutorials</a></li>
                            <li><a href="#">Blog</a></li>
                        </ul>
                    </div>
                    <div class="footer-col">
                        <h4>Company</h4>
                        <ul>
                            <li><a href="#">About</a></li>
                            <li><a href="#">Contact</a></li>
                            <li><a href="#">Support</a></li>
                        </ul>
                    </div>
                </div>
            </div>
            <div class="footer-bottom">
                <p>&copy; <?php echo date('Y'); ?> Lightspeed. All rights reserved.</p>
                <p>Powered by Lightspeed</p>
            </div>
        </div>
    </footer>

    <script src="assets/js/main.js"></script>
</body>
</html>

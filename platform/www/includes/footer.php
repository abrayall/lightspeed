    <footer class="site-footer">
        <div class="container">
            <div class="footer-content">
                <div class="footer-brand">
                    <a href="<?php echo $basePath ?? ''; ?>" class="logo">
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
                            <li><a href="<?php echo $basePath ?? ''; ?>#features">Features</a></li>
                        </ul>
                    </div>
                    <div class="footer-col">
                        <h4>Documentation</h4>
                        <ul>
                            <li><a href="<?php echo $basePath ?? ''; ?>docs/">Getting Started</a></li>
                            <li><a href="<?php echo $basePath ?? ''; ?>docs/#cmd-init">CLI Commands</a></li>
                            <li><a href="<?php echo $basePath ?? ''; ?>docs/#site-properties">Configuration</a></li>
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

    <script src="<?php echo $basePath ?? ''; ?>assets/js/main.js"></script>
<?php if (isset($extraJs)): ?>
<?php echo $extraJs; ?>
<?php endif; ?>
</body>
</html>

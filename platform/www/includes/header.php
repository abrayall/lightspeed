<?php
require_once('lightspeed/version.php');
require_once('lightspeed/cms.php');

$_CMS = cms('https://velocity-server-f2oqe.ondigitalocean.app');
?>
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title><?php echo $pageTitle ?? 'Lightspeed | Fast Website Hosting & Development Platform'; ?></title>
    <link rel="icon" type="image/svg+xml" href="data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 48 48'><circle cx='24' cy='24' r='22' fill='%23deb841'/><circle cx='24' cy='24' r='10' fill='white'/><circle cx='24' cy='24' r='7' fill='%23deb841'/></svg>">
    <link rel="stylesheet" href="<?php echo $basePath ?? ''; ?>assets/css/style.css">
<?php if (isset($extraCss)): ?>
<?php foreach ($extraCss as $css): ?>
    <link rel="stylesheet" href="<?php echo $css; ?>">
<?php endforeach; ?>
<?php endif; ?>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700;800&family=JetBrains+Mono:wght@400;500&display=swap" rel="stylesheet">

    <style>
        <?= $_CMS->get('default', 'theme', 'css') ?>
    </style>
</head>
<body class="<?php echo $bodyClass ?? ''; ?>">
    <header class="site-header">
        <div class="container header-inner">
            <a href="<?php echo $basePath ?? ''; ?>" class="logo">
                <svg class="logo-icon" width="28" height="28" viewBox="0 0 48 48">
                    <circle cx="24" cy="24" r="22" fill="var(--accent)"/>
                    <circle cx="24" cy="24" r="10" fill="white"/>
                    <circle cx="24" cy="24" r="7" fill="var(--accent)"/>
                </svg>
                <span class="logo-text">Lightspeed</span>
            </a>
            <nav class="main-nav">
                <ul class="nav-menu">
                    <li><a href="<?php echo $basePath ?? ''; ?>#features">Features</a></li>
                    <li><a href="<?php echo $basePath ?? ''; ?>#how-it-works">Get Started</a></li>
                    <li><a href="<?php echo $basePath ?? ''; ?>docs/"<?php echo ($currentPage ?? '') === 'docs' ? ' class="active"' : ''; ?>>Documentation</a></li>
                </ul>
                <a href="<?php echo $basePath ?? ''; ?>#how-it-works" class="btn btn-primary">Develop Now</a>
            </nav>
            <button class="mobile-menu-toggle" aria-label="Toggle menu">
                <span></span>
                <span></span>
                <span></span>
            </button>
        </div>
    </header>

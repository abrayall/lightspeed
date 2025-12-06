<?php
$pageTitle = 'Error | Lightspeed';
$bodyClass = 'error-page';
include 'includes/header.php';
?>

    <main class="error-content">
        <div class="container">
            <div class="error-message">
                <h1>Oops, we hit an issue</h1>
                <p>Something went wrong. Please try again later.</p>
                <a href="<?php echo $basePath ?? ''; ?>" class="btn btn-primary">Go Home</a>
            </div>
        </div>
    </main>

<?php include 'includes/footer.php'; ?>

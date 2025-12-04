<?php
/**
 * Lightspeed Mail
 *
 * Simple email sending using PHPMailer with SMTP settings from site.properties
 */

require_once __DIR__ . '/phpmailer/Exception.php';
require_once __DIR__ . '/phpmailer/PHPMailer.php';
require_once __DIR__ . '/phpmailer/SMTP.php';
require_once __DIR__ . '/site.php';

use PHPMailer\PHPMailer\PHPMailer;
use PHPMailer\PHPMailer\Exception;

class Mail {
    private PHPMailer $mailer;
    private Site $site;

    public function __construct(?Site $site = null) {
        $this->site = $site ?? site();
        $this->mailer = new PHPMailer(true);
        $this->configure();
    }

    private function configure(): void {
        $this->mailer->isSMTP();
        $this->mailer->Host = $this->site->get('smtp.host', 'localhost');
        $this->mailer->Port = $this->site->getInt('smtp.port', 587);
        $this->mailer->SMTPSecure = $this->site->get('smtp.secure', 'tls');

        $username = $this->site->getEncrypted('smtp.username');
        $password = $this->site->getEncrypted('smtp.password');

        if ($username !== '' && $password !== '') {
            $this->mailer->SMTPAuth = true;
            $this->mailer->Username = $username;
            $this->mailer->Password = $password;
        }

        $fromEmail = $this->site->get('smtp.from', $this->site->get('email'));
        $fromName = $this->site->get('smtp.from.name', $this->site->name());

        if ($fromEmail !== '') {
            $this->mailer->setFrom($fromEmail, $fromName);
        }
    }

    public function to(string $email, string $name = ''): self {
        $this->mailer->addAddress($email, $name);
        return $this;
    }

    public function cc(string $email, string $name = ''): self {
        $this->mailer->addCC($email, $name);
        return $this;
    }

    public function bcc(string $email, string $name = ''): self {
        $this->mailer->addBCC($email, $name);
        return $this;
    }

    public function replyTo(string $email, string $name = ''): self {
        $this->mailer->addReplyTo($email, $name);
        return $this;
    }

    public function subject(string $subject): self {
        $this->mailer->Subject = $subject;
        return $this;
    }

    public function body(string $body, bool $isHtml = false): self {
        if ($isHtml) {
            $this->mailer->isHTML(true);
            $this->mailer->Body = $body;
        } else {
            $this->mailer->Body = $body;
        }
        return $this;
    }

    public function html(string $html, string $altText = ''): self {
        $this->mailer->isHTML(true);
        $this->mailer->Body = $html;
        if ($altText !== '') {
            $this->mailer->AltBody = $altText;
        }
        return $this;
    }

    public function attach(string $path, string $name = ''): self {
        $this->mailer->addAttachment($path, $name);
        return $this;
    }

    public function send(): bool {
        try {
            return $this->mailer->send();
        } catch (Exception $e) {
            return false;
        }
    }

    public function error(): string {
        return $this->mailer->ErrorInfo;
    }
}

function mailer(): Mail {
    return new Mail();
}

function send_mail(string $to, string $subject, string $body, bool $isHtml = false): bool {
    return mailer()
        ->to($to)
        ->subject($subject)
        ->body($body, $isHtml)
        ->send();
}

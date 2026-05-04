<?php
/**
 * Lightspeed Database
 *
 * Simple MySQL database access using PDO with connection settings from environment variables.
 *
 * Environment variables:
 *   DB_HOST     - Database host
 *   DB_PORT     - Database port (default: 3306)
 *   DB_NAME     - Database name
 *   DB_USER     - Database username
 *   DB_PASSWORD  - Database password
 *   DB_SSL      - SSL mode: REQUIRED or empty (default: empty)
 */

class Database {
    private static ?Database $instance = null;
    private ?PDO $pdo = null;

    private string $host;
    private int $port;
    private string $name;
    private string $user;
    private string $password;
    private bool $ssl;

    /**
     * Get the singleton Database instance
     */
    public static function instance(): Database {
        if (self::$instance === null) {
            self::$instance = new Database();
        }
        return self::$instance;
    }

    public function __construct(
        ?string $host = null,
        ?int $port = null,
        ?string $name = null,
        ?string $user = null,
        ?string $password = null,
        ?bool $ssl = null
    ) {
        $this->host = $host ?? getenv('DB_HOST') ?: 'localhost';
        $this->port = $port ?? (int)(getenv('DB_PORT') ?: 3306);
        $this->name = $name ?? getenv('DB_NAME') ?: '';
        $this->user = $user ?? getenv('DB_USER') ?: '';
        $this->password = $password ?? getenv('DB_PASSWORD') ?: '';
        $this->ssl = $ssl ?? (strtoupper(getenv('DB_SSL') ?: '') === 'REQUIRED');
    }

    /**
     * Get the PDO connection (lazy-initialized)
     */
    public function connection(): PDO {
        if ($this->pdo === null) {
            $dsn = sprintf('mysql:host=%s;port=%d', $this->host, $this->port);
            if ($this->name !== '') {
                $dsn .= ';dbname=' . $this->name;
            }
            $dsn .= ';charset=utf8mb4';

            $options = [
                PDO::ATTR_ERRMODE => PDO::ERRMODE_EXCEPTION,
                PDO::ATTR_DEFAULT_FETCH_MODE => PDO::FETCH_ASSOC,
                PDO::ATTR_EMULATE_PREPARES => false,
            ];

            if ($this->ssl) {
                $options[PDO::MYSQL_ATTR_SSL_VERIFY_SERVER_CERT] = false;
                $options[PDO::MYSQL_ATTR_SSL_CA] = '';
            }

            $this->pdo = new PDO($dsn, $this->user, $this->password, $options);
        }
        return $this->pdo;
    }

    /**
     * Execute a query and return all results
     */
    public function query(string $sql, array $params = []): array {
        $stmt = $this->connection()->prepare($sql);
        $stmt->execute($params);
        return $stmt->fetchAll();
    }

    /**
     * Execute a query and return a single row
     */
    public function queryOne(string $sql, array $params = []): ?array {
        $stmt = $this->connection()->prepare($sql);
        $stmt->execute($params);
        $row = $stmt->fetch();
        return $row !== false ? $row : null;
    }

    /**
     * Execute a statement (INSERT, UPDATE, DELETE) and return affected rows
     */
    public function execute(string $sql, array $params = []): int {
        $stmt = $this->connection()->prepare($sql);
        $stmt->execute($params);
        return $stmt->rowCount();
    }

    /**
     * Insert a row and return the last insert ID
     */
    public function insert(string $table, array $data): string {
        $columns = implode(', ', array_keys($data));
        $placeholders = implode(', ', array_fill(0, count($data), '?'));
        $sql = "INSERT INTO {$table} ({$columns}) VALUES ({$placeholders})";
        $this->execute($sql, array_values($data));
        return $this->connection()->lastInsertId();
    }

    /**
     * Test the connection
     */
    public function test(): bool {
        try {
            $this->connection()->query('SELECT 1');
            return true;
        } catch (PDOException $e) {
            return false;
        }
    }

    /**
     * Get connection error (if test fails)
     */
    public function error(): string {
        try {
            $this->connection();
            return '';
        } catch (PDOException $e) {
            return $e->getMessage();
        }
    }

    /**
     * Get server version info
     */
    public function version(): string {
        return $this->connection()->getAttribute(PDO::ATTR_SERVER_VERSION);
    }
}

/**
 * Helper function to get the database instance
 */
function db(): Database {
    return Database::instance();
}

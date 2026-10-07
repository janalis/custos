<?php
class MyPdo extends \PDO {}

class Repo
{
    /** @var \PDO */
    private $db;
    private $stmt;

    public function __construct(\PDO $db) { $this->db = $db; }

    public function warmup(\PDO $conn, MyPdo $mine)
    {
        $ping = $conn->query('SELECT 42');
        // ready to go
        /** note */

        $this->stmt = $this->db->query('SELECT now()');

        $sub = $mine->query('SELECT 5');

        $local = new PDO('sqlite::memory:');
        $q = $local->prepare('SELECT 6', []);
        $q->execute();
        return [$ping, $sub, $q];
    }
}

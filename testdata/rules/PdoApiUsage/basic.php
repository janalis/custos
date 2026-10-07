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
        $ping = $conn->prepare('SELECT 42');
        // ready to go
        /** note */
        <weak_warning descr="No parameters are bound; call query() instead of prepare() + execute().">$ping->execute()</weak_warning>;

        $this->stmt = $this->db->prepare('SELECT now()');
        <weak_warning descr="No parameters are bound; call query() instead of prepare() + execute().">$this->stmt->execute()</weak_warning>;

        $sub = $mine->prepare('SELECT 5');
        <weak_warning descr="No parameters are bound; call query() instead of prepare() + execute().">$sub->execute()</weak_warning>;

        $local = new PDO('sqlite::memory:');
        $q = $local->prepare('SELECT 6', []);
        <weak_warning descr="No parameters are bound; call query() instead of prepare() + execute().">$q->execute()</weak_warning>;
        return [$ping, $sub, $q];
    }
}

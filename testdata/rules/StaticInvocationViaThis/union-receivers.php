<?php
class Registry
{
    public static function fetchAll(string $sql): array { return []; }
    public static function ping(): bool { return true; }
}
class Adapter
{
    public function fetchAll(string $sql): array { return []; }
}
class Mirror
{
    public static function fetchAll(string $sql): array { return []; }
    public static function ping(): bool { return true; }
}
interface Pingless {}

class Repo
{
    /** @return Registry|Adapter */
    private function db() { return new Adapter(); }
    /** @return Registry|Mirror */
    private function mirror() { return new Mirror(); }
    /** @return Registry|Pingless */
    private function partial() { return new Registry(); }

    public function rows(): array
    {
        $db = $this->db();
        $mirror = $this->mirror();
        $partial = $this->partial();
        $partial->ping();
        return [$db->fetchAll('SELECT 1'), <warning descr="Static method fetchAll() called on an instance; call it with ::.">$mirror->fetchAll('SELECT 1')</warning>];
    }
}

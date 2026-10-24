<?php
class Cluster
{
    protected $pool = [];
    private static $slaves = [];
    private $count = 0;

    public function add($id, $connection)
    {
        $this->pool[$id] = $connection;
        self::$slaves[] = $connection;
        $this->count++;
    }

    public function random()
    {
        // Other methods fill the properties: their defaults prove nothing.
        if (!$this->pool) {
            return null;
        }
        $a = $this->pool[array_rand($this->pool)];
        $b = array_rand(self::$slaves);
        $c = array_rand(['x', 'y'], $this->count);
        $d = <error descr="Request a count within the array size.">array_rand([])</error>;
        return [$a, $b, $c, $d];
    }
}

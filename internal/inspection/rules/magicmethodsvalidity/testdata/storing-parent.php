<?php
class Storing
{
    public $db;
    public $conf;
    public function __construct($db, $conf = null)
    {
        $this->db = $db;
        $this->conf = $conf;
    }
}
class SetsBoth extends Storing
{
    public function __construct($db)
    {
        $this->db = $db;
        $this->conf = null;
        $this->extra = 1;
    }
}
class SetsOne extends Storing
{
    public function <error descr="__construct does not call Storing::__construct().">__construct</error>($db)
    {
        $this->db = $db;
    }
}
class Promoting
{
    public function __construct(protected $db, $flag = false)
    {
        $this->flag = $flag;
    }
}
class FromPromoting extends Promoting
{
    public function __construct(protected $db, $flag = true)
    {
        $this->flag = $flag;
    }
}
class FromPromotingPartly extends Promoting
{
    public function <error descr="__construct does not call Promoting::__construct().">__construct</error>($db)
    {
        $this->db = $db;
    }
}
class Working
{
    public function __construct($db)
    {
        $this->db = $db;
        $this->connect();
    }
    private function connect() {}
}
class FromWorking extends Working
{
    public function <error descr="__construct does not call Working::__construct().">__construct</error>($db)
    {
        $this->db = $db;
    }
}
class Computing
{
    public function __construct($db)
    {
        $this->db = strtolower($db);
    }
}
class FromComputing extends Computing
{
    public function <error descr="__construct does not call Computing::__construct().">__construct</error>($db)
    {
        $this->db = $db;
    }
}
class Other
{
    public function __construct($a) { $this->a = $a; }
    public function __clone() { $this->a = clone $this->a; }
}
class FromOther extends Other
{
    public function __construct($a) { $this->a = $a; }
}

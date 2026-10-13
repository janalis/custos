<?php
class OnlyPublic
{
    public $a;
    public function __construct() { $this->a = 1; }
}

class Untouched
{
    private $a;
    public function __construct() { echo 1; }
}

// The trait lives in another file: its methods cannot be scanned.
class UsesExternal
{
    use ExternalTrait;
    private $conn;
    public function __construct() { $this->conn = 1; }
}

// An unknown trait is ignored.
class UsesUnknown
{
    use MissingTrait;
    private $conn;
    public function __construct() { <weak_warning descr="Private property is only used in the constructor; likely dead code.">$this->conn</weak_warning> = 1; }
}

class Statics
{
    private $seen;
    private $kept;
    private $other;
    public function __construct($x)
    {
        <weak_warning descr="Private property is only used in the constructor; likely dead code.">self::$seen</weak_warning> = 1;
        static::$kept = 1;
        parent::$other = 1;
        // an unresolvable receiver may be this class (D3)
        <weak_warning descr="Private property is only used in the constructor; likely dead code.">$x::$seen</weak_warning> = 2;
        Statics::$kept = 3;
    }
    public function kept() { return static::$kept; }
}

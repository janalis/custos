<?php
// Properties read by a computed name or all at once may be any of them.
class CmInfo
{
    private $indent;
    private $name;

    public function __construct($mod)
    {
        $this->indent = $mod->indent;
        $this->name = $mod->name;
    }

    public function __get($prop)
    {
        return $this->$prop;
    }
}

class Braced
{
    private $a;

    public function __construct($a) { $this->a = $a; }

    public function get($k) { return $this->{'_' . $k}; }
}

class Exported
{
    private $a;

    public function __construct($a) { $this->a = $a; }

    public function export() { return get_object_vars($this); }
}

class Iterated
{
    private $a;

    public function __construct($a) { $this->a = $a; }

    public function all() { $out = []; foreach ($this as $k => $v) { $out[$k] = $v; } return $out; }
}

class Casted
{
    private $a;

    public function __construct($a) { $this->a = $a; }

    public function toArray() { return (array) $this; }
}

trait Dumps
{
    public function dump() { return get_object_vars($this); }
}

class WithTrait
{
    use Dumps;

    private $a;

    public function __construct($a) { $this->a = $a; }
}

// Still reported: other objects' dynamic properties and unrelated calls.
class Other
{
    private $a;

    public function __construct($a) { <weak_warning descr="Private property is only used in the constructor; likely dead code.">$this->a</weak_warning> = $a; }

    public function read($o, $k) { return [$o->$k, get_object_vars($o), (array) $o, (int) $this]; }
}
class Fetcher
{
    private $permanentUrl;

    public function __construct($url, $redirects = 5)
    {
        if ($redirects > 0 && $url === 'moved') {
            $this->permanentUrl = $url;
            $this->__construct('target', $redirects - 1);
        }
    }
}

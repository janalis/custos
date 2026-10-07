<?php
namespace App;

// Parent declared in another file: defaults compared as text, doc comments
// unavailable, D10 decided by the parent's @return doc type.
class Local extends \Lib\Remote
{
    public function other($a = 2) { return parent::other($a); }
    public function loud($a) { parent::loud($a); }
    /** @return mixed */
    public function documented($a) { return parent::documented($a); }
}

class Base
{
    public function __construct(protected $id, $name) {}
    public function plain($a) { return $a; }
    public function withDefault($a = 1) { return $a; }
    public function noDefault($a) { return $a; }
    /** @param int $a */
    public function tagged($a) { return $a; }
    public function nested($a) { $f = function () { return 1; }; if ($a) { return; } echo $f(); }
    public function early($a) { if ($a) { return 1; } echo 2; }
}

class Child extends Base
{
    public function __construct($id, $name) { parent::__construct($id, $name); }
    public function plain($a) { return self::plain($a); }
    public function withDefault($a = 2) { return parent::withDefault($a); }
    public function noDefault($a = 1) { return parent::noDefault($a); }
    /** @param string $a */
    public function tagged($a) { return parent::tagged($a); }
    public function early($a) { parent::early($a); }
    public function extra($a) { return parent::extra($a, 1); }
}

$anon = new class extends Base {
    public function plain($a) { return parent::plain($a); }
};

class Orphan
{
    public function plain($a) { return parent::plain($a); }
}

<?php
namespace Illuminate\Database\Eloquent;

class Model { public static function query() {} }

class Invoice extends Model {
    public static function recent() {}
    public function scope() { return $this->query(); }
    public function latest() { return <warning descr="Static method recent() called through $this; use self::recent().">$this</warning>->recent(); }
}

<?php
namespace Illuminate\Database\Eloquent;

class Model { public static function query() {} }

class Invoice extends Model {
    public function scope() { return <warning descr="Static method query() called through $this; use self::query().">$this</warning>->query(); }
}

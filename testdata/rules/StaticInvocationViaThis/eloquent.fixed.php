<?php
namespace Illuminate\Database\Eloquent;

class Model { public static function query() {} }

class Invoice extends Model {
    public static function recent() {}
    public function scope() { return $this->query(); }
    public function latest() { return static::recent(); }
}

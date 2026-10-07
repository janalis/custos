<?php
namespace Illuminate\Database\Eloquent;

class Model { public static function query() {} }

class Invoice extends Model {
    public function scope() { return self::query(); }
}

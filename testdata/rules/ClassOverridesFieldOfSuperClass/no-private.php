<?php
class Model { private $secret; protected $table; }
class Invoice extends Model {
    public $secret;
    protected <weak_warning descr="Property 'table' is already declared in \Model; drop this re-declaration.">$table</weak_warning>;
}

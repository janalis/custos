<?php
class Model {
    protected $table;
    protected $visible;
    protected static $registry;
    const KIND = 'model';
}

/** @property string $virtual */
class Invoice extends Model {
    public $visible;
    protected static $registry;
    const KIND = 'invoice';
    public $virtual;
    public function __construct(protected $table = 'x') {}
}

class InvoiceTest extends Model {
    protected $table;
}

class Entity { protected $uuid; }
class Order extends Entity {
    /** @Column(type="guid") */
    protected $uuid;
}

class Locked {
    protected $state;
    final protected function __construct($seed = 1) {}
}
class Door extends Locked {
    protected $state;
}
class Hidden { protected $state; private function __construct() {} }
class Mid extends Hidden {}
class Leaf extends Mid { protected $state; }
class Free extends Unknown { protected $state; }

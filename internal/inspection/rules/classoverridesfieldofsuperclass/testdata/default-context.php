<?php
namespace App\Base;

const PREFIX = 'base';

class Limits { const MAX = 10; }

class Model {
    const TABLE = 'models';
    const KEY = 'id';
    protected $table = self::TABLE;
    protected $key = self::KEY;
    protected $origin = __CLASS__;
    protected $space = __NAMESPACE__;
    protected $kind = self::class;
    protected $max = Limits::MAX;
    protected $absolute = \App\Base\Limits::MAX;
    protected $prefix = PREFIX;
    protected $flags = [true, false, null, \PHP_INT_SIZE];
    protected $root = parent::ROOT;
    protected $dynamic = self::{'KEY'};
    protected $late = static::KEY;
}

class Invoice extends Model {
    const TABLE = 'invoices';
    // Same text, another constant: self:: is bound to the declaring class.
    protected $table = self::TABLE;
    protected $origin = __CLASS__;
    protected $space = __NAMESPACE__;
    protected $kind = self::class;
    protected $root = parent::ROOT;
    protected $dynamic = self::{'KEY'};
    protected $late = static::KEY;
    // Both declarations reach the same constant in the same scope.
    protected <weak_warning descr="Property 'key' is already declared in \App\Base\Model; drop this re-declaration.">$key</weak_warning> = self::KEY;
    protected <weak_warning descr="Property 'max' is already declared in \App\Base\Model; drop this re-declaration.">$max</weak_warning> = Limits::MAX;
    protected <weak_warning descr="Property 'absolute' is already declared in \App\Base\Model; drop this re-declaration.">$absolute</weak_warning> = \App\Base\Limits::MAX;
    protected <weak_warning descr="Property 'prefix' is already declared in \App\Base\Model; drop this re-declaration.">$prefix</weak_warning> = PREFIX;
    protected <weak_warning descr="Property 'flags' is already declared in \App\Base\Model; drop this re-declaration.">$flags</weak_warning> = [true, false, null, \PHP_INT_SIZE];
}

namespace App\Other;

const PREFIX = 'other';

class Limits { const MAX = 20; }

class Report extends \App\Base\Model {
    // Unqualified names resolve in this namespace, not the parent's.
    protected $max = Limits::MAX;
    protected $prefix = PREFIX;
    protected <weak_warning descr="Property 'key' is already declared in \App\Base\Model; drop this re-declaration.">$key</weak_warning> = self::KEY;
    protected <weak_warning descr="Property 'absolute' is already declared in \App\Base\Model; drop this re-declaration.">$absolute</weak_warning> = \App\Base\Limits::MAX;
    protected <weak_warning descr="Property 'flags' is already declared in \App\Base\Model; drop this re-declaration.">$flags</weak_warning> = [true, false, null, \PHP_INT_SIZE];
}

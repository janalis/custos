<?php
namespace Shop;

class Limits { const MAX = 50; }

class Books extends Catalog {
    // The parent's file may import another Limits: not provably the same.
    protected $max = Limits::MAX;
    protected <weak_warning descr="Property 'perPage' is already declared in \Shop\Catalog; drop this re-declaration.">$perPage</weak_warning> = self::PAGE;
    protected <weak_warning descr="Property 'absolute' is already declared in \Shop\Catalog; drop this re-declaration.">$absolute</weak_warning> = \Shop\Limits::MAX;
}

<?php
namespace Shop;

use Vendor\Defaults as Limits;

class Catalog {
    const PAGE = 20;
    protected $perPage = self::PAGE;
    protected $max = Limits::MAX;
    protected $absolute = \Shop\Limits::MAX;
}

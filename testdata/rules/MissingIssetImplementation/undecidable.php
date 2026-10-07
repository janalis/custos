<?php

namespace Shop\Catalog;

use Acme\Absent\Record;

interface Entity
{
    public function id(): int;
}

abstract class Model {}

#[\AllowDynamicProperties]
class Module
{
    public string $name = '';
}

class Theme extends Module {}

class Failure {}

#[Tagged]
class Product {}

class Price {}

enum Size
{
    case Small;
}

function checks(Entity $e, Model $m, Theme $t, Record $r, Size $s, ?Product $p): array
{
    return [
        !empty($e->in_preview),
        isset($m->cache),
        !empty($t->sub_themes),
        isset($r->id),
        isset($s->label),
        isset(<error descr="\Shop\Catalog\Product has no __isset(); this isset/empty check is always false.">$p->sku</error>),
    ];
}

/** @param object|Failure $api */
function api($api): bool
{
    return !empty($api->homepage);
}

/** @param array|Failure $data */
function data($data): bool
{
    return isset($data->homepage);
}

/** @param \stdClass|Failure $row */
function row($row): bool
{
    return isset($row->homepage);
}

/** @param Product|Record $item */
function item($item): bool
{
    return isset($item->homepage);
}

/** @param Product|Price $item */
function both($item): bool
{
    return isset(<error descr="\Shop\Catalog\Price has no __isset(); this isset/empty check is always false.">$item->homepage</error>);
}

class Oops extends \Exception {}

function oops(Oops $e): bool
{
    return isset(<error descr="\Shop\Catalog\Oops has no __isset(); this isset/empty check is always false.">$e->detail</error>);
}

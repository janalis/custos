<?php
namespace Shop;

final class Order
{
    public array $lines = [];
    public ?Order $parent = null;
}

function iterate(Order $order, array $list): void
{
    // The iterated subject of foreach is read.
    foreach ($order-><error descr="Fix the property name: Shop\Order declares no $itemz.">itemz</error> as $line) {}
    foreach ($order-><error descr="Fix the property name: Shop\Order declares no $itemz.">itemz</error>[0] as $key => $line) {}
    // Iterating by reference writes it; the key and value are targets.
    foreach ($order->byRef as &$line) {}
    foreach ($list as $order->key => $order->value) {}
    // A read elsewhere is still reported: iterating creates nothing.
    echo $order-><error descr="Fix the property name: Shop\Order declares no $itemz.">itemz</error>;
}

function destructure(Order $order, array $list): void
{
    // A key of a destructuring target is read; its value is written.
    [$order-><error descr="Fix the property name: Shop\Order declares no $k.">k</error> => $value] = $list;
    [$order-><error descr="Fix the property name: Shop\Order declares no $k.">k</error>[0] => [$value]] = $list;
    ['k' => $order->assigned] = $list;
    echo $order-><error descr="Fix the property name: Shop\Order declares no $k.">k</error>, $order->assigned;
}

function remove(Order $order): void
{
    // unset() of anything below a property never warns about it.
    unset($order->meta->x, $order->meta->y['z'], $order->parent->meta->x);
    unset(($order->meta)->z, $order->tags['x']);
    // Unsetting below a property creates nothing.
    echo $order-><error descr="Fix the property name: Shop\Order declares no $meta.">meta</error>;
}

function update(Order $order): void
{
    // Compound assignments and increments read the property first.
    $order-><error descr="Fix the property name: Shop\Order declares no $note.">note</error> .= 'x';
    $order-><error descr="Fix the property name: Shop\Order declares no $count.">count</error>++;
    --$order-><error descr="Fix the property name: Shop\Order declares no $count.">count</error>;
    $order-><error descr="Fix the property name: Shop\Order declares no $sums.">sums</error>[0] += 1;
    $order-><error descr="Fix the property name: Shop\Order declares no $sums.">sums</error>[0][1]++;
    // A plain or `??=` assignment does not read it.
    $order->fresh = 1;
    $order->fresh .= 'x';
    $order->lazy ??= 1;
    $order->cells[0] ??= 1;
    $order->lines[0] .= 'x';
}

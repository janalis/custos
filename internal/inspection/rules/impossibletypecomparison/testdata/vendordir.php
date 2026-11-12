<?php
namespace Shop;

use Acme\Input;
use Acme\Output;
use Deps\Gate;

final class Order {}

// The project installs its dependencies into deps/ (Composer's
// config.vendor-dir): a directory named vendor holds project code.
function vendorDir(Input $in, Output $out, Order $order, Gate $gate): array
{
    return [
        <error descr="This comparison is always false: Acme\Input and Acme\Output share no value.">$in === $out</error>,
        <error descr="This comparison is always false: Shop\Order and Acme\Input share no value.">$order === $in</error>,
        // A final class of the project never extends a dependency class.
        <error descr="This comparison is always false: Shop\Order and Deps\Gate share no value.">$order === $gate</error>,
        <error descr="This comparison is always false: Acme\Input and Deps\Gate share no value.">$in === $gate</error>,
        $gate == $order,
    ];
}

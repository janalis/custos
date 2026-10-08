<?php

namespace Catalog {
    class Product {}
    class Variant {}
    function helper() {}
}

namespace Storefront {
    use Catalog\{Product, Variant as V, function helper};
    use function Catalog\helper as assist;
    use const Catalog\LIMIT;

    $name = 'X';
    return [
        Product::class,
        product::class,
        V::class,
        v::class,
        Product::{$name},
    ];
}

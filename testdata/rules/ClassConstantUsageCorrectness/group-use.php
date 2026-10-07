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
        <error descr="Letter case of the class name differs from its declaration; ::class will return the wrong string.">product</error>::class,
        V::class,
        <error descr="Letter case of the class name differs from its declaration; ::class will return the wrong string.">v</error>::class,
        Product::{$name},
    ];
}

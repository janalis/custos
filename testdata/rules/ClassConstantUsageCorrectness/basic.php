<?php

namespace Catalog {
    class Product {}
    class Variant {}
}

namespace Storefront {
    use Catalog\product;
    use Catalog\Variant as Option;

    return [
        <error descr="Letter case of the class name differs from its declaration; ::class will return the wrong string.">Product</error>::class,
        <error descr="Letter case of the class name differs from its declaration; ::class will return the wrong string.">OPTION</error>::class,
        Option::class,
        <error descr="Letter case of the class name differs from its declaration; ::class will return the wrong string.">Widgets\banner</error>::class,
        Widgets\Banner::class,
        Unknown\Thing::class,
        missing::class,
    ];
}

namespace Storefront\Widgets {
    use Catalog\Product;

    class Banner {}

    return [
        Product::class,
        <error descr="Letter case of the class name differs from its declaration; ::class will return the wrong string.">product</error>::class,
        banner::class,
    ];
}

namespace Checkout {
    use Storefront\Widgets as W;
    use Storefront\Widgets\Banner as Promo;

    class Cart
    {
        public function names($obj)
        {
            return [
                W\Banner::class,
                <error descr="Letter case of the class name differs from its declaration; ::class will return the wrong string.">W\BANNER</error>::class,
                <error descr="Letter case of the class name differs from its declaration; ::class will return the wrong string.">\arrayobject</error>::class,
                \ArrayObject::class,
                Promo::class,
                self::class,
                static::class,
                parent::class,
                $obj::class,
                SELF::class,
            ];
        }
    }
}

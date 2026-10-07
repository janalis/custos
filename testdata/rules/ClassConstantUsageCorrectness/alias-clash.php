<?php
namespace Catalog {
    class Product {}
    class Variant {}
}

namespace Shop {
    use Catalog\Product as Item;
    use Catalog\Variant as ITEM;

    return [Item::class];
}

<?php
use PHPUnit\Framework\Attributes\DataProvider;
class AttributeDatasets {
    public function values() { return [[1]]; }
    #[DataProvider(methodName: 'values')]
    public function <weak_warning descr="Give the provider's datasets string keys.">testValues</weak_warning>() {}
}

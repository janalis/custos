<?php
class Order
{
    /**
     * @param bool     $isFinal
     * @param callable $onDone
     */
    public function __construct(private $isFinal, private $onDone, private $other)
    {
    }

    public function isFinal(): bool { return $this->isFinal; }

    public function <weak_warning descr="A callable property with this name exists; rename the method (for example with a get/is/has prefix).">onDone</weak_warning>(): void {}

    public function <weak_warning descr="A property with this name exists and its type is unknown; rename the method or type the property.">other</weak_warning>(): void {}
}

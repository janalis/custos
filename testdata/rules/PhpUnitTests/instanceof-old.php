<?php
class OrphanTest
{
    public function testOrphan($x)
    {
        <weak_warning descr="Use 'assertInstanceOf()' instead.">$this->assertTrue($x instanceof parent)</weak_warning>;
    }
}

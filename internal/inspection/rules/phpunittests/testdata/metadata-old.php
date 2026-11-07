<?php
use PHPUnit\Framework\Attributes\DataProvider;
use PHPUnit\Framework\Attributes\CoversClass;
#[CoversClass('Missing')]
class OldVersion {
    #[DataProvider('missing')]
    public function testOld() {}
}

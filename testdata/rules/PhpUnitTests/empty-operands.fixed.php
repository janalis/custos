<?php
class ArchiveTest
{
    /**
     * @param int[] $counts
     */
    public function testEmpty(array $ids, ?string $name, \ZipArchive $zip, \Countable $bag, $any, array $counts, int|float|bool|null $n, iterable $it)
    {
        self::assertNotEmpty($ids);
        $this->assertEmpty($name, 'name');
        $this->assertEmpty($name);
        self::assertNotTrue(empty($zip));
        self::assertNotEmpty($counts, 'm', 3);
        $this->assertEmpty($n);
        $this->assertNotEmpty($counts);
        $this->assertNotFalse(empty($any));
        $this->assertNotTrue(empty($zip));
        $this->assertFalse(empty($bag));
        $this->assertTrue(empty($any));
        $this->assertTrue(empty($it));
    }
}

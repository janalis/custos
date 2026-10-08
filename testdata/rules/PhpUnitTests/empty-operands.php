<?php
class ArchiveTest
{
    /**
     * @param int[] $counts
     */
    public function testEmpty(array $ids, ?string $name, \SplObjectStorage $zip, \Countable $bag, $any, array $counts, int|float|bool|null $n, iterable $it)
    {
        <weak_warning descr="Use 'assertNotEmpty()' instead.">self::assertTrue(!empty($ids))</weak_warning>;
        <weak_warning descr="Use 'assertEmpty()' instead.">$this->assertFalse(!empty($name), 'name')</weak_warning>;
        <weak_warning descr="Use 'assertEmpty()' instead.">$this->assertTrue(empty($name))</weak_warning>;
        <weak_warning descr="Use 'assertNotTrue()' instead.">self::assertTrue(!empty($zip))</weak_warning>;
        <weak_warning descr="Use 'assertNotEmpty()' instead.">self::assertTrue(!(empty($counts)), 'm', 3)</weak_warning>;
        <weak_warning descr="Use 'assertEmpty()' instead.">$this->assertNotFalse(empty($n))</weak_warning>;
        <weak_warning descr="Use 'assertNotEmpty()' instead.">$this->assertNotTrue(empty($counts))</weak_warning>;
        <weak_warning descr="Use 'assertNotFalse()' instead.">$this->assertFalse(!empty($any))</weak_warning>;
        $this->assertNotTrue(empty($zip));
        $this->assertFalse(empty($bag));
        $this->assertTrue(empty($any));
        $this->assertTrue(empty($it));
    }
}

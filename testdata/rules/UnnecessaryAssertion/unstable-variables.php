<?php

namespace Shop;

abstract class LedgerTest
{
    abstract protected function flush(): void;

    public function testUnstable(array $rows)
    {
        $result = $this->flush();
        foreach ($rows as $row) {
            $result .= $row;
        }
        $this->assertNull($result);
    }
}

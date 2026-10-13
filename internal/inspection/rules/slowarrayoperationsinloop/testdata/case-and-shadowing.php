<?php
namespace App {
    function array_merge(...$arrays) { return []; }
    function count($v) { return 0; }

    function collect(array $batches) {
        $all = [];
        foreach ($batches as $batch) {
            $all = array_merge($all, $batch);
        }
        for ($i = 0; $i < count($batches); $i++) {}
        return $all;
    }
}

namespace {
    function gather(array $batches) {
        $all = [];
        foreach ($batches as $batch) {
            $all = <error descr="'array_merge(...)' inside a loop re-copies the accumulator each time; merge once after the loop.">Array_Merge($all, $batch)</error>;
        }
        return $all;
    }
}

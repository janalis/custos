<?php
function byValue($value, $other = null)
{
    return $value;
}

function byReference($source, &$target, &...$rest)
{
    $target = $source;
}

// Nothing in these loops defines the variable for a later iteration.
function neverDefined(array $rows)
{
    foreach ($rows as $row) {
        if (isset(<error descr="Variable '$passed' is not defined in this scope.">$passed</error>)) {
            echo 1;
        }
        byValue($passed);
        byValue(other: $passed, value: $row);
        byReference($passed, $row);
        byReference(...);
        echo strlen($passed), new \stdClass;
    }
    foreach ($rows as $row) {
        if (isset(<error descr="Variable '$object' is not defined in this scope.">$object</error>)) {
            echo 1;
        }
        $object->id = $row;
    }
    foreach ($rows as $row) {
        if (isset(<error descr="Variable '$gone' is not defined in this scope.">$gone</error>)) {
            echo 1;
        }
        unset($gone);
        $rows[$gone] = [$gone];
    }
    foreach ($rows as $row) {
        if (isset(<error descr="Variable '$inner' is not defined in this scope.">$inner</error>)) {
            echo 1;
        }
        $set = function () use ($row) {
            $inner = $row;
        };
        $get = fn () => $inner = $row;
        function nestedHelper()
        {
            $inner = 1;
        }
        new class {
            public function run()
            {
                $inner = 1;
            }
        };
    }
}

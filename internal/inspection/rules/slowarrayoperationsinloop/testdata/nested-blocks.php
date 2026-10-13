<?php
function gather(array $groups, bool $strict): array
{
    $rows = [];
    foreach ($groups as $kind => $group) {
        switch ($kind) {
            case 'extra':
                $rows = array_merge($rows, $group);
                break;
        }
        try {
            if ($strict) {
                $rows = array_merge($rows, $group);
            }
        } catch (\Exception $e) {
        }
        foreach ($group as $item) {
            if ($item === null) {
                {
                    $rows = array_merge($rows, [$item]);
                }
            }
        }
    }
    foreach ($groups as $group) {
        try {
            $rows = array_merge($rows, $group);
            throw new \RuntimeException('done');
        } finally {
        }
    }
    foreach ($groups as $group) {
        if ($strict) {
            try {
                $rows = array_merge($rows, $group);
            } finally {
            }
        }
    }
    foreach ($groups as $group) {
        $rows = array_merge($rows, $group);
        throw new \LogicException('once');
    }
    foreach ($groups as $group) {
        try {
            $rows = <error descr="'array_merge(...)' inside a loop re-copies the accumulator each time; merge once after the loop.">array_merge($rows, $group)</error>;
        } finally {
        }
        {
            $rows = <error descr="'array_replace(...)' inside a loop re-copies the accumulator each time; merge once after the loop.">array_replace($rows, $group)</error>;
        }
    }
    return $rows;
}

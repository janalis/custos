<?php
namespace Lists {
    function array_is_list(array $a): bool { return false; }

    function check(array $rows): array
    {
        return [
            \array_is_list($rows),
            !\array_is_list($rows),
        ];
    }
}

namespace Imported {
    use function Lists\array_is_list;

    function check(array $rows): bool
    {
        return \array_is_list($rows);
    }
}

namespace Plain {
    function check(array $rows): bool
    {
        return array_is_list($rows);
    }
}

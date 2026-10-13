<?php
namespace App\Export {
    // No qualified global constant here (\true does not count): bare names.
    function dump(array $v): string|false
    {
        return \true ? json_encode($v, JSON_THROW_ON_ERROR) : false;
    }
}

namespace App\Shadow {
    const JSON_THROW_ON_ERROR = 0;

    // A bare name would reach the constant above.
    function dump(array $v): string|false
    {
        return json_encode($v, \JSON_THROW_ON_ERROR);
    }
}

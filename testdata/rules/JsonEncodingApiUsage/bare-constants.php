<?php
namespace App\Export {
    // No qualified global constant here (\true does not count): bare names.
    function dump(array $v): string|false
    {
        return \true ? <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($v)</weak_warning> : false;
    }
}

namespace App\Shadow {
    const JSON_THROW_ON_ERROR = 0;

    // A bare name would reach the constant above.
    function dump(array $v): string|false
    {
        return <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($v)</weak_warning>;
    }
}

<?php
namespace Ids {
    $a = <error descr="Pass more_entropy = true to uniqid() to reduce collisions.">UniqId('x')</error>;
    $b = Array_Map(<error descr="Pass more_entropy = true to uniqid() to reduce collisions.">'UNIQID'</error>, ['a']);
}

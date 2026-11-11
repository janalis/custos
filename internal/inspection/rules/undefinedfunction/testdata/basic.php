<?php
namespace App;

use function Lib\Text\slug;

function local_helper() {}

<error descr="Fix the name or install the package declaring str_contians().">str_contians</error>('haystack', 'needle');
<error descr="Fix the name or install the package declaring strlenn().">strlenn</error>('x');
<error descr="Fix the name or install the package declaring App\Util\slugify().">Util\slugify</error>('Title');
<error descr="Fix the name or install the package declaring Lib\Text\slug().">slug</error>('Title');
<error descr="Fix the name or install the package declaring Other\run().">\Other\run</error>();

local_helper();
strlen('x');
\strlen('x');
array_map(null, []);

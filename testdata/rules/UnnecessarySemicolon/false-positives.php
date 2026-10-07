<?php
declare(ticks=1);
declare(strict_types=1) ;

for ($i = 0; $i < 3; $i++) ;
foreach ($items as $item) ;
while (next($items)) ;
do ; while (false);
if ($ready) ;
elseif ($late) ;
else if ($later) ;
else ;

function plain() {
    return;
}
?>
<p><?= $title; $title = null; ?></p>
<p><?= $title ?></p>
<p><?php show($title); ?></p>
<p><?php echo $title; ?></p>

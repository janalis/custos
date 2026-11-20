<?php // presets: per
if ($a):
  echo 1;
elseif ($b):
      echo 2;
else:
  echo 3;
endif;
foreach ($l as $v):
echo $v;
endforeach;
while ($x):
  $x--;
endwhile;
for (;;):
  break;
endfor;
switch ($a):
  case 1:
    break;
endswitch;
declare(ticks=1):
  tick();
enddeclare;
declare(ticks=1) { tick(); }

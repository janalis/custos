<?php
namespace Negative0 { preg_split('/(,)/',$s,-1,PREG_SPLIT_DELIM_CAPTURE); }
namespace Negative1 { preg_split('/,/',$s); }
namespace Negative2 { preg_split('/,/',$s,-1,0); }
namespace Negative3 { preg_split($pattern,$s,-1,PREG_SPLIT_DELIM_CAPTURE); }
namespace Negative4 { preg_split('/(?|,)/',$s,-1,PREG_SPLIT_DELIM_CAPTURE); }
namespace Negative5 { preg_split('/,/',$s,-1,$flags); }
namespace Negative6 { trim($s); }
namespace Negative7 { preg_split('/,/',$s,-1,PREG_SPLIT_NO_EMPTY); }

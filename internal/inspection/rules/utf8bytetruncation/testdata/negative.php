<?php
namespace Negative0 { substr('ñame',0,2); }
namespace Negative1 { substr('name',0,1); }
namespace Negative2 { substr($s,0,1); }
namespace Negative3 { substr('ñame',$start,1); }
namespace Negative4 { substr('ñame',20,1); }
namespace Negative5 { substr('ñame',-100,2); }
namespace Negative6 { substr('ñame',0,-3); }
namespace Negative7 { substr('ñame',0,$length); }
namespace Negative8 { substr('ñame',0,-100); }
namespace Negative9 { strlen('ñ'); }
namespace Negative10 { substr('ñame',5,-100); }

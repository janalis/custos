<?php
namespace Negative0 { preg_match('/x/','ax',$m,PREG_OFFSET_CAPTURE);mb_substr('ax',$m[0][1],1); }
namespace Negative1 { preg_match('/x/u','ñx',$m);mb_substr('ñx',$m[0][1],1); }
namespace Negative2 { preg_match('/x/u','ñx',$m,PREG_OFFSET_CAPTURE);substr('ñx',$m[0][1],1); }
namespace Negative3 { preg_match('/x/u','ñx',$m,PREG_OFFSET_CAPTURE);mb_substr('other',$m[0][1],1); }
namespace Negative4 { mb_substr('ñx',0,1); }
namespace Negative5 { mb_substr('ñx',$m[0][0],1); }
namespace Negative6 { mb_substr('ñx',$m[1],1); }
namespace Negative7 { $m=other();mb_substr('ñx',$m[0][1],1); }
namespace Negative8 { other();mb_substr('ñx',$m[0][1],1); }
namespace Negative9 { preg_match('/x/u','ñx',$other,PREG_OFFSET_CAPTURE);mb_substr('ñx',$m[0][1],1); }
namespace Negative10 { mb_substr('ñx',items()[0][1],1); }
namespace Negative11 { preg_match('/x/u','öx',$m,PREG_OFFSET_CAPTURE);mb_substr('ñx',$m[0][1],1); }

namespace Audit0 { preg_match('/x/u','ñx',$m,PREG_OFFSET_CAPTURE);mb_substr('ñx',$m[0][1],1,'8bit'); }

namespace Audit1 { preg_match('/x/u','ñx',$m,PREG_OFFSET_CAPTURE);mb_substr('ñx',$m[0][1],1,$encoding); }

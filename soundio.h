#ifndef _C_SOUNDIO_H_
#define _C_SOUNDIO_H_

#include <soundio/soundio.h>

extern void setSoundIoCallback(struct SoundIo *);
extern void setInStreamCallback(struct SoundIoInStream *);
extern void setOutStreamCallback(struct SoundIoOutStream *);

#endif // _C_SOUNDIO_H_

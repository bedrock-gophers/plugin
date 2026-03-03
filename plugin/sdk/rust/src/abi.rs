pub const EVENT_PLUGIN_COMMAND: u16 = 39;

#[derive(Default)]
pub struct Encoder {
    buf: Vec<u8>,
}

impl Encoder {
    pub fn with_capacity(capacity: usize) -> Self {
        Self {
            buf: Vec::with_capacity(capacity),
        }
    }

    pub fn u8(&mut self, value: u8) {
        self.buf.push(value);
    }

    pub fn bool(&mut self, value: bool) {
        self.u8(if value { 1 } else { 0 });
    }

    pub fn u32(&mut self, value: u32) {
        self.buf.extend_from_slice(&value.to_le_bytes());
    }

    pub fn bytes(&mut self, value: &[u8]) {
        self.u32(value.len() as u32);
        self.buf.extend_from_slice(value);
    }

    pub fn string(&mut self, value: &str) {
        self.bytes(value.as_bytes());
    }

    pub fn into_bytes(self) -> Vec<u8> {
        self.buf
    }
}

pub struct Decoder<'a> {
    data: &'a [u8],
    off: usize,
    ok: bool,
}

impl<'a> Decoder<'a> {
    pub fn new(data: &'a [u8]) -> Self {
        Self {
            data,
            off: 0,
            ok: true,
        }
    }

    pub fn ok(&self) -> bool {
        self.ok
    }

    pub fn u8(&mut self) -> u8 {
        if self.off + 1 > self.data.len() {
            self.ok = false;
            return 0;
        }
        let value = self.data[self.off];
        self.off += 1;
        value
    }

    pub fn bool(&mut self) -> bool {
        self.u8() == 1
    }

    pub fn u32(&mut self) -> u32 {
        if self.off + 4 > self.data.len() {
            self.ok = false;
            return 0;
        }
        let mut buf = [0u8; 4];
        buf.copy_from_slice(&self.data[self.off..self.off + 4]);
        self.off += 4;
        u32::from_le_bytes(buf)
    }

    pub fn bytes(&mut self) -> Vec<u8> {
        let size = self.u32() as usize;
        if self.off + size > self.data.len() {
            self.ok = false;
            return Vec::new();
        }
        let out = self.data[self.off..self.off + size].to_vec();
        self.off += size;
        out
    }

    pub fn string(&mut self) -> String {
        String::from_utf8(self.bytes()).unwrap_or_default()
    }
}

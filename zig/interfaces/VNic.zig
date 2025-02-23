pub const VNic = struct {
    this:*VNic,
    pub fn Start(this: *VNic) !void {
        return this.Start(this);
    }
    pub fn Shutdown(this: *VNic) !void {
        return this.Shutdown(this);
    }
};
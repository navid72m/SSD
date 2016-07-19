# -*- coding: utf-8 -*-
"""
Created on Sun Jul 17 11:31:56 2016

@author: navid
"""
import time
import struct
class Record:
    record=bytearray()
    checksum=0
    timestamp=0
    block=0
    raw_blk=0
    worker_id=0
    op_cnt=0
    seed=0
    marker=bytearray()
    
    def __init__(self):
        self.set_time()
        self.set_marker()
        record=timestamp+marker
        
        print "Record"        
        print record
        
    
    def set_time(self):
        global timestamp
        tmp_time=time.clock();
        print "tmp_time"
        print tmp_time
        timestamp=struct.pack('f', tmp_time)
        print "timestamp"
        print timestamp
        t=struct.unpack('f',timestamp)
        print "tmp_time2"
        print t
        
    def set_marker(self):
        global marker
        marker=b'\x23\x23\x23'
        print marker
    


r=Record()
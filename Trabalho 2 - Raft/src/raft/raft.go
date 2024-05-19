package raft

//
// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
// rf = Make(...)
//   create a new Raft server.
// rf.Start(command interface{}) (index, term, isleader)
//   start agreement on a new log entry
// rf.GetState() (term, isLeader)
//   ask a Raft for its current term, and whether it thinks it is leader
// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester)
//   in the same server.
//

import "fmt"
import "sync"
import "labrpc"
import "math/rand"
import "time"

// import "bytes"
// import "encoding/gob"



type Role int

const (
	Follower Role = iota
	Candidate
	Leader
)

type AppendEntriesArgs struct {
	Term int // leader's term
	LeaderId int // so follower can redirect clients
}

type AppendEntriesResult struct {
	Term int // currentTerm, for leader to update itself
	Success bool // true if follower contained entry matching prevLogIndex and prevLogTerm
}

//
// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make().
//

type ApplyMsg struct {
	Index       int
	Command     interface{}
	UseSnapshot bool   // ignore for lab2; only used in lab3
	Snapshot    []byte // ignore for lab2; only used in lab3
}

//
// A Go object implementing a single Raft peer.
//
type Raft struct {
	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]

	// Your data here (2A, 2B, 2C).
	// Look at the paper's Figure 2 for a description of what
	// state a Raft server must maintain.
	currentTerm int // latest term server has seen
	votedFor int // candidateId that received vote in current term
	role Role // role of server (0: follower, 1: candidate, 2: leader)
	timer int64 // timer for election timeout
	startTime int64 // start time of timer
	votesReceived int // number of votes received in election
	leaderId int // leader's id
	isDead bool // server is dead or alive
}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {

	var term int
	var isleader bool
	// Your code here (2A).
	term = rf.currentTerm
	isleader = rf.role == Leader

	return term, isleader // return currentTerm and whether this server believes it is the leader
}

//
// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
//
func (rf *Raft) persist() {
	// Your code here (2C).
	// Example:
	// w := new(bytes.Buffer)
	// e := gob.NewEncoder(w)
	// e.Encode(rf.xxx)
	// e.Encode(rf.yyy)
	// data := w.Bytes()
	// rf.persister.SaveRaftState(data)
}

//
// restore previously persisted state.
//
func (rf *Raft) readPersist(data []byte) {
	// Your code here (2C).
	// Example:
	// r := bytes.NewBuffer(data)
	// d := gob.NewDecoder(r)
	// d.Decode(&rf.xxx)
	// d.Decode(&rf.yyy)
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
}




//
// example RequestVote RPC arguments structure.
// field names must start with capital letters!
//
type RequestVoteArgs struct {
	Term int // candidate's term
	CandidateId int // candidate requesting vote
	LastLogIndex int // index of candidate's last log entry
	LastLogTerm int // term of candidate's last log entry
}

//
// example RequestVote RPC reply structure.
// field names must start with capital letters!
//
type RequestVoteReply struct {
	Term int // currentTerm, for candidate to update itself
	VoteGranted bool // true means candidate received vote
}

//
// example RequestVote RPC handler.
//
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	if args.Term < rf.currentTerm {
		reply.Term = rf.currentTerm // return currentTerm so candidate can update itself
		reply.VoteGranted = false // false because term is less than currentTerm
		return
	}

	rf.votedFor = args.CandidateId // vote for candidate
	rf.currentTerm = args.Term // update currentTerm
	rf.role = Follower // change role to follower
	rf.startTime = currentTime() // reset start time of timer

	reply.Term = args.Term // return currentTerm so candidate can update itself
	reply.VoteGranted = true // true because term is greater than or equal to currentTerm
}

//
// example code to send a RequestVote RPC to a server.
// server is the index of the target server in rf.peers[].
// expects RPC arguments in args.
// fills in *reply with RPC reply, so caller should
// pass &reply.
// the types of the args and reply passed to Call() must be
// the same as the types of the arguments declared in the
// handler function (including whether they are pointers).
//
// The labrpc package simulates a lossy network, in which servers
// may be unreachable, and in which requests and replies may be lost.
// Call() sends a request and waits for a reply. If a reply arrives
// within a timeout interval, Call() returns true; otherwise
// Call() returns false. Thus Call() may not return for a while.
// A false return can be caused by a dead server, a live server that
// can't be reached, a lost request, or a lost reply.
//
// Call() is guaranteed to return (perhaps after a delay) *except* if the
// handler function on the server side does not return.  Thus there
// is no need to implement your own timeouts around Call().
//
// look at the comments in ../labrpc/labrpc.go for more details.
//
// if you're having trouble getting RPC to work, check that you've
// capitalized all field names in structs passed over RPC, and
// that the caller passes the address of the reply struct with &, not
// the struct itself.
//
func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}


//
// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election.
//
// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// term. the third return value is true if this server believes it is
// the leader.
//
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	index := -1
	term := -1
	isLeader := true

	// Your code here (2B).


	return index, term, isLeader
}

//
// the tester calls Kill() when a Raft instance won't
// be needed again. you are not required to do anything
// in Kill(), but it might be convenient to (for example)
// turn off debug output from this instance.
//
func (rf *Raft) Kill() {
	// Your code here, if desired.
	rf.isDead = true // set isDead to true
}

//
// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
//
func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me
	rf.currentTerm = 0
	rf.votedFor = -1
	rf.role = Follower
	rf.timer = randomTimeFollower()
	rf.startTime = currentTime()
	rf.votesReceived = 0
	rf.leaderId = -1
	rf.isDead = false

	// Your initialization code here (2A, 2B, 2C).
	go rf.routine() // start routine for server

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())


	return rf
}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesResult) {
	if args.Term < rf.currentTerm {
		reply.Term = rf.currentTerm // return currentTerm so leader can update itself
		reply.Success = false // false because term is less than currentTerm
		return
	}

	rf.currentTerm = args.Term // update currentTerm
	reply.Term = args.Term // return currentTerm so leader can update itself
	reply.Success = true // true because term is greater than or equal to currentTerm

	if(rf.role != Follower) {
		rf.role = Follower // change role to follower
	}

	rf.startTime = currentTime() // reset start time of timer
}

func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesResult) bool {
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	return ok
}

func (rf *Raft) broadcastHeartbeat() {
	for {
		if rf.role != Leader || rf.isDead {
			break
		}
		var appendEntriesArgs *AppendEntriesArgs = &AppendEntriesArgs{Term: rf.currentTerm, LeaderId: rf.leaderId}
		for i := 0; i < len(rf.peers); i++ {
			if i != rf.me {
				var appendEntriesResult *AppendEntriesResult = &AppendEntriesResult{Term: -1, Success: false}
				var ok = rf.sendAppendEntries(i, appendEntriesArgs, appendEntriesResult)
				if(ok && !appendEntriesResult.Success) {
					rf.currentTerm = appendEntriesResult.Term // update currentTerm
					rf.role = Follower // change role to follower
					rf.startTime = currentTime() // reset start time of timer
					break
				}
			}
		}
	}
}

func (rf *Raft) routine() {
	for {
		var currTime = currentTime()
		if rf.isDead {
			return
		}
		if rf.role == Follower {
			if currTime - rf.startTime > rf.timer {
				rf.role = Candidate // change role to candidate
				rf.currentTerm++ // increment currentTerm
				rf.votedFor = rf.me // vote for self
				rf.votesReceived = 1 // vote for self
				rf.startTime = currentTime() // reset start time of timer
				go rf.startElection() // start broadcasting request vote messages
			}
		} else if rf.role == Leader {
			go rf.broadcastHeartbeat()
		} else if rf.role == Candidate {
			if currTime - rf.startTime > rf.timer {
				//rf.startElection()
			}
		}
	}
}

func (rf *Raft) startElection() {
	fmt.Println("Starting election")
	rf.role = Candidate // change role to candidate
	rf.currentTerm++ // increment currentTerm
	rf.votedFor = rf.me // vote for self
	rf.votesReceived = 1 // vote for self
	rf.startTime = currentTime() // reset start time of timer
	go rf.broadcastRequestVote() // start broadcasting request vote messages
}

func (rf *Raft) broadcastRequestVote() {
	fmt.Println("Broadcasting request vote")
    for i := 0; i < len(rf.peers); i++ {
		if i != rf.me {
			fmt.Println("Sending request vote from ", rf.me, "to ", i)
			var requestVoteArgs *RequestVoteArgs = &RequestVoteArgs{Term: rf.currentTerm, CandidateId: rf.me, LastLogIndex: 0, LastLogTerm: 0}
			var requestVoteReply *RequestVoteReply = &RequestVoteReply{Term: -1, VoteGranted: false}
			var ok = rf.sendRequestVote(i, requestVoteArgs, requestVoteReply)
			if(ok && requestVoteReply.VoteGranted) {
				fmt.Println("Vote granted from ", i, " to ", rf.me)
				rf.votesReceived++ // increment votes received
				if(rf.votesReceived > len(rf.peers)/2) { // if majority votes received
					rf.role = Leader // change role to leader
					rf.votedFor = -1 // reset votedFor
					rf.votesReceived = 0 // reset votesReceived
					rf.leaderId = rf.me // set leaderId
					fmt.Println("Leader elected: ", rf.me)
					go rf.broadcastHeartbeat() // start sending heartbeat messages
					return
				}
			}
		}
	}
}

func randomTimeFollower() int64 {
	return int64(500 + rand.Intn(500)) // random time between 150ms and 300ms
}

func currentTime() int64 {
	return int64(time.Now().UnixNano()/1000000) // current time in milliseconds
}
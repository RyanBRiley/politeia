package decredplugin

import "encoding/json"

// Plugin settings, kinda doesn;t go here but for now it is fine
const (
	Version                  = "1"
	ID                       = "decred"
	CmdAuthorizeVote         = "authorizevote"
	CmdStartVote             = "startvote"
	CmdVoteDetails           = "votedetails"
	CmdVoteSummary           = "votesummary"
	CmdLoadVoteResults       = "loadvoteresults"
	CmdBallot                = "ballot"
	CmdBestBlock             = "bestblock"
	CmdNewComment            = "newcomment"
	CmdLikeComment           = "likecomment"
	CmdCensorComment         = "censorcomment"
	CmdGetComment            = "getcomment"
	CmdGetComments           = "getcomments"
	CmdGetNumComments        = "getnumcomments"
	CmdProposalVotes         = "proposalvotes"
	CmdCommentLikes          = "commentlikes"
	CmdProposalCommentsLikes = "proposalcommentslikes"
	CmdInventory             = "inventory"
	CmdTokenInventory        = "tokeninventory"
	MDStreamAuthorizeVote    = 13 // Vote authorization by proposal author
	MDStreamVoteBits         = 14 // Vote bits and mask
	MDStreamVoteSnapshot     = 15 // Vote tickets and start/end parameters

	VoteDurationMin = 2016 // Minimum vote duration (in blocks)
	VoteDurationMax = 4032 // Maximum vote duration (in blocks)

	// Authorize vote actions
	AuthVoteActionAuthorize = "authorize" // Authorize a proposal vote
	AuthVoteActionRevoke    = "revoke"    // Revoke a proposal vote authorization
)

// CastVote is a signed vote.
type CastVote struct {
	Token     string `json:"token"`     // Proposal ID
	Ticket    string `json:"ticket"`    // Ticket ID
	VoteBit   string `json:"votebit"`   // Vote bit that was selected, this is encode in hex
	Signature string `json:"signature"` // Signature of Token+Ticket+VoteBit
}

// Ballot is a batch of votes that are sent to the server.
type Ballot struct {
	Votes []CastVote `json:"votes"`
}

// EncodeCastVotes encodes CastVotes into a JSON byte slice.
func EncodeBallot(b Ballot) ([]byte, error) {
	return json.Marshal(b)
}

// DecodeCastVotes decodes a JSON byte slice into a CastVotes.
func DecodeBallot(payload []byte) (*Ballot, error) {
	var b Ballot

	err := json.Unmarshal(payload, &b)
	if err != nil {
		return nil, err
	}

	return &b, nil
}

// CastVoteReply contains the signature or error to a cast vote command.
type CastVoteReply struct {
	ClientSignature string `json:"clientsignature"` // Signature that was sent in
	Signature       string `json:"signature"`       // Signature of the ClientSignature
	Error           string `json:"error"`           // Error if something wen't wrong during casting a vote
}

// EncodeCastVoteReply encodes CastVoteReply into a JSON byte slice.
func EncodeCastVoteReply(cvr CastVoteReply) ([]byte, error) {
	return json.Marshal(cvr)
}

// DecodeBallotReply decodes a JSON byte slice into a CastVotes.
func DecodeCastVoteReply(payload []byte) (*CastVoteReply, error) {
	var cvr CastVoteReply

	err := json.Unmarshal(payload, &cvr)
	if err != nil {
		return nil, err
	}

	return &cvr, nil
}

// BallotReply is a reply to a batched list of votes.
type BallotReply struct {
	Receipts []CastVoteReply `json:"receipts"`
}

// EncodeCastVoteReplies encodes CastVotes into a JSON byte slice.
func EncodeBallotReply(br BallotReply) ([]byte, error) {
	return json.Marshal(br)
}

// DecodeBallotReply decodes a JSON byte slice into a CastVotes.
func DecodeBallotReply(payload []byte) (*BallotReply, error) {
	var br BallotReply

	err := json.Unmarshal(payload, &br)
	if err != nil {
		return nil, err
	}

	return &br, nil
}

// VoteOption describes a single vote option.
type VoteOption struct {
	Id          string `json:"id"`          // Single unique word identifying vote (e.g. yes)
	Description string `json:"description"` // Longer description of the vote.
	Bits        uint64 `json:"bits"`        // Bits used for this option
}

// Vote represents the vote options for vote that is identified by its token.
type Vote struct {
	Token            string       `json:"token"`            // Token that identifies vote
	Mask             uint64       `json:"mask"`             // Valid votebits
	Duration         uint32       `json:"duration"`         // Duration in blocks
	QuorumPercentage uint32       `json:"quorumpercentage"` // Percent of eligible votes required for quorum
	PassPercentage   uint32       `json:"passpercentage"`   // Percent of total votes required to pass
	Options          []VoteOption `json:"options"`          // Vote option
}

// EncodeVote encodes Vote into a JSON byte slice.
func EncodeVote(v Vote) ([]byte, error) {
	return json.Marshal(v)
}

// DecodeVote decodes a JSON byte slice into a Vote.
func DecodeVote(payload []byte) (*Vote, error) {
	var v Vote

	err := json.Unmarshal(payload, &v)
	if err != nil {
		return nil, err
	}

	return &v, nil
}

// AuthorizeVote is an MDStream that is used to indicate that a proposal has
// been finalized and is ready to be voted on.  The signature and public
// key are from the proposal author.  The author can revoke a previously sent
// vote authorization by setting the Action field to revoke.
const VersionAuthorizeVote = 1

type AuthorizeVote struct {
	// Generated by decredplugin
	Version   uint   `json:"version"`   // Version of this structure
	Receipt   string `json:"receipt"`   // Server signature of client signature
	Timestamp int64  `json:"timestamp"` // Received UNIX timestamp

	// Generated by client
	Action    string `json:"action"`    // Authorize or revoke
	Token     string `json:"token"`     // Proposal censorship token
	Signature string `json:"signature"` // Signature of token+version+action
	PublicKey string `json:"publickey"` // Pubkey used for signature
}

// EncodeAuthorizeVote encodes AuthorizeVote into a JSON byte slice.
func EncodeAuthorizeVote(av AuthorizeVote) ([]byte, error) {
	return json.Marshal(av)
}

// DecodeAuthorizeVote decodes a JSON byte slice into an AuthorizeVote.
func DecodeAuthorizeVote(payload []byte) (*AuthorizeVote, error) {
	var av AuthorizeVote
	err := json.Unmarshal(payload, &av)
	if err != nil {
		return nil, err
	}
	return &av, nil
}

// AuthorizeVoteReply returns the authorize vote action that was executed and
// the receipt for the action.  The receipt is the server side signature of
// AuthorizeVote.Signature.
type AuthorizeVoteReply struct {
	Action        string `json:"action"`        // Authorize or revoke
	RecordVersion string `json:"recordversion"` // Version of record files
	Receipt       string `json:"receipt"`       // Server signature of client signature
	Timestamp     int64  `json:"timestamp"`     // Received UNIX timestamp
}

// EncodeAuthorizeVote encodes AuthorizeVoteReply into a JSON byte slice.
func EncodeAuthorizeVoteReply(avr AuthorizeVoteReply) ([]byte, error) {
	return json.Marshal(avr)
}

// DecodeAuthorizeVoteReply decodes a JSON byte slice into a AuthorizeVoteReply.
func DecodeAuthorizeVoteReply(payload []byte) (*AuthorizeVoteReply, error) {
	var avr AuthorizeVoteReply
	err := json.Unmarshal(payload, &avr)
	if err != nil {
		return nil, err
	}
	return &avr, nil
}

// StartVote instructs the plugin to commence voting on a proposal with the
// provided vote bits.
const VersionStartVote = 1

type StartVote struct {
	// decred plugin only data
	Version uint `json:"version"` // Version of this structure

	PublicKey string `json:"publickey"` // Key used for signature.
	Vote      Vote   `json:"vote"`      // Vote + options
	Signature string `json:"signature"` // Signature of Votehash
}

// EncodeStartVoteencodes StartVoteinto a JSON byte slice.
func EncodeStartVote(v StartVote) ([]byte, error) {
	return json.Marshal(v)
}

// DecodeVotedecodes a JSON byte slice into a StartVote.
func DecodeStartVote(payload []byte) (*StartVote, error) {
	var sv StartVote

	err := json.Unmarshal(payload, &sv)
	if err != nil {
		return nil, err
	}

	return &sv, nil
}

// StartVoteReply is the reply to StartVote.
const VersionStartVoteReply = 1

type StartVoteReply struct {
	// decred plugin only data
	Version uint `json:"version"` // Version of this structure

	// Shared data
	StartBlockHeight string   `json:"startblockheight"` // Block height
	StartBlockHash   string   `json:"startblockhash"`   // Block hash
	EndHeight        string   `json:"endheight"`        // Height of vote end
	EligibleTickets  []string `json:"eligibletickets"`  // Valid voting tickets
}

// EncodeStartVoteReply encodes StartVoteReply into a JSON byte slice.
func EncodeStartVoteReply(v StartVoteReply) ([]byte, error) {
	return json.Marshal(v)
}

// DecodeVoteReply decodes a JSON byte slice into a StartVoteReply.
func DecodeStartVoteReply(payload []byte) (*StartVoteReply, error) {
	var v StartVoteReply

	err := json.Unmarshal(payload, &v)
	if err != nil {
		return nil, err
	}

	return &v, nil
}

// VoteDetails is used to retrieve the voting period details for a record.
type VoteDetails struct {
	Token string `json:"token"` // Censorship token
}

// EncodeVoteDetails encodes VoteDetails into a JSON byte slice.
func EncodeVoteDetails(vd VoteDetails) ([]byte, error) {
	return json.Marshal(vd)
}

// DecodeVoteDetails decodes a JSON byte slice into a VoteDetails.
func DecodeVoteDetails(payload []byte) (*VoteDetails, error) {
	var vd VoteDetails

	err := json.Unmarshal(payload, &vd)
	if err != nil {
		return nil, err
	}

	return &vd, nil
}

// VoteDetailsReply is the reply to VoteDetails.
type VoteDetailsReply struct {
	AuthorizeVote  AuthorizeVote  `json:"authorizevote"`  // Vote authorization
	StartVote      StartVote      `json:"startvote"`      // Vote ballot
	StartVoteReply StartVoteReply `json:"startvotereply"` // Start vote snapshot
}

// EncodeVoteDetailsReply encodes VoteDetailsReply into a JSON byte slice.
func EncodeVoteDetailsReply(vdr VoteDetailsReply) ([]byte, error) {
	return json.Marshal(vdr)
}

// DecodeVoteReply decodes a JSON byte slice into a VoteDetailsReply.
func DecodeVoteDetailsReply(payload []byte) (*VoteDetailsReply, error) {
	var vdr VoteDetailsReply

	err := json.Unmarshal(payload, &vdr)
	if err != nil {
		return nil, err
	}

	return &vdr, nil
}

type VoteResults struct {
	Token string `json:"token"` // Censorship token
}

type VoteResultsReply struct {
	StartVote StartVote  `json:"startvote"` // Original ballot
	CastVotes []CastVote `json:"castvotes"` // All votes
}

// EncodeVoteResults encodes VoteResults into a JSON byte slice.
func EncodeVoteResults(v VoteResults) ([]byte, error) {
	return json.Marshal(v)
}

// DecodeVoteResults decodes a JSON byte slice into a VoteResults.
func DecodeVoteResults(payload []byte) (*VoteResults, error) {
	var v VoteResults

	err := json.Unmarshal(payload, &v)
	if err != nil {
		return nil, err
	}

	return &v, nil
}

// EncodeVoteResultsReply encodes VoteResults into a JSON byte slice.
func EncodeVoteResultsReply(v VoteResultsReply) ([]byte, error) {
	return json.Marshal(v)
}

// DecodeVoteResultsReply decodes a JSON byte slice into a VoteResults.
func DecodeVoteResultsReply(payload []byte) (*VoteResultsReply, error) {
	var v VoteResultsReply

	err := json.Unmarshal(payload, &v)
	if err != nil {
		return nil, err
	}

	return &v, nil
}

// VoteSummary requests a summary of a proposal vote. This includes certain
// voting period parameters and a summary of the vote results.
type VoteSummary struct {
	Token string `json:"token"` // Censorship token
}

// EncodeVoteSummary encodes VoteSummary into a JSON byte slice.
func EncodeVoteSummary(v VoteSummary) ([]byte, error) {
	return json.Marshal(v)
}

// DecodeVoteSummary decodes a JSON byte slice into a VoteSummary.
func DecodeVoteSummary(payload []byte) (*VoteSummary, error) {
	var v VoteSummary

	err := json.Unmarshal(payload, &v)
	if err != nil {
		return nil, err
	}

	return &v, nil
}

// VoteOptionResult describes a vote option and the total number of votes that
// have been cast for this option.
type VoteOptionResult struct {
	ID          string `json:"id"`          // Single unique word identifying vote (e.g. yes)
	Description string `json:"description"` // Longer description of the vote.
	Bits        uint64 `json:"bits"`        // Bits used for this option
	Votes       uint64 `json:"votes"`       // Number of votes cast for this option
}

// VoteSummaryReply is the reply to the VoteSummary command and returns certain
// voting period parameters as well as a summary of the vote results.
type VoteSummaryReply struct {
	Authorized          bool               `json:"authorized"`          // Vote is authorized
	EndHeight           string             `json:"endheight"`           // End block height
	EligibleTicketCount int                `json:"eligibleticketcount"` // Number of eligible tickets
	QuorumPercentage    uint32             `json:"quorumpercentage"`    // Percent of eligible votes required for quorum
	PassPercentage      uint32             `json:"passpercentage"`      // Percent of total votes required to pass
	Results             []VoteOptionResult `json:"results"`             // Vote results
}

// EncodeVoteSummaryReply encodes VoteSummary into a JSON byte slice.
func EncodeVoteSummaryReply(v VoteSummaryReply) ([]byte, error) {
	return json.Marshal(v)
}

// DecodeVoteSummaryReply decodes a JSON byte slice into a VoteSummaryReply.
func DecodeVoteSummaryReply(payload []byte) (*VoteSummaryReply, error) {
	var v VoteSummaryReply

	err := json.Unmarshal(payload, &v)
	if err != nil {
		return nil, err
	}

	return &v, nil
}

// Comment is the structure that describes the full server side content.  It
// includes server side meta-data as well. Note that the receipt is the server
// side.
type Comment struct {
	// Data generated by client
	Token     string `json:"token"`     // Censorship token
	ParentID  string `json:"parentid"`  // Parent comment ID
	Comment   string `json:"comment"`   // Comment
	Signature string `json:"signature"` // Client Signature of Token+ParentID+Comment
	PublicKey string `json:"publickey"` // Pubkey used for Signature

	// Metadata generated by decred plugin
	CommentID   string `json:"commentid"`   // Comment ID
	Receipt     string `json:"receipt"`     // Server signature of the client Signature
	Timestamp   int64  `json:"timestamp"`   // Received UNIX timestamp
	TotalVotes  uint64 `json:"totalvotes"`  // Total number of up/down votes
	ResultVotes int64  `json:"resultvotes"` // Vote score
	Censored    bool   `json:"censored"`    // Has this comment been censored
}

// EncodeComment encodes Comment into a JSON byte slice.
func EncodeComment(c Comment) ([]byte, error) {
	return json.Marshal(c)
}

// DecodeComment decodes a JSON byte slice into a Comment
func DecodeComment(payload []byte) (*Comment, error) {
	var c Comment

	err := json.Unmarshal(payload, &c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

// NewComment sends a comment from a user to a specific proposal.  Note that
// the user is implied by the session.
type NewComment struct {
	Token     string `json:"token"`     // Censorship token
	ParentID  string `json:"parentid"`  // Parent comment ID
	Comment   string `json:"comment"`   // Comment
	Signature string `json:"signature"` // Signature of Token+ParentID+Comment
	PublicKey string `json:"publickey"` // Pubkey used for Signature
}

// EncodeNewComment encodes NewComment into a JSON byte slice.
func EncodeNewComment(nc NewComment) ([]byte, error) {
	return json.Marshal(nc)
}

// DecodeNewComment decodes a JSON byte slice into a NewComment
func DecodeNewComment(payload []byte) (*NewComment, error) {
	var nc NewComment

	err := json.Unmarshal(payload, &nc)
	if err != nil {
		return nil, err
	}

	return &nc, nil
}

// NewCommentReply returns the metadata generated by decred plugin for the new
// comment.
type NewCommentReply struct {
	CommentID string `json:"commentid"` // Comment ID
	Receipt   string `json:"receipt"`   // Server signature of the client Signature
	Timestamp int64  `json:"timestamp"` // Received UNIX timestamp
}

// EncodeNewCommentReply encodes NewCommentReply into a JSON byte slice.
func EncodeNewCommentReply(ncr NewCommentReply) ([]byte, error) {
	return json.Marshal(ncr)
}

// DecodeNewCommentReply decodes a JSON byte slice into a NewCommentReply.
func DecodeNewCommentReply(payload []byte) (*NewCommentReply, error) {
	var ncr NewCommentReply

	err := json.Unmarshal(payload, &ncr)
	if err != nil {
		return nil, err
	}

	return &ncr, nil
}

// LikeComment records an up or down vote from a user on a comment.
type LikeComment struct {
	Token     string `json:"token"`     // Censorship token
	CommentID string `json:"commentid"` // Comment ID
	Action    string `json:"action"`    // Up or downvote (1, -1)
	Signature string `json:"signature"` // Client Signature of Token+CommentID+Action
	PublicKey string `json:"publickey"` // Pubkey used for Signature

	// Only used on disk
	Receipt   string `json:"receipt,omitempty"`   // Signature of Signature
	Timestamp int64  `json:"timestamp,omitempty"` // Received UNIX timestamp
}

// EncodeLikeComment encodes LikeComment into a JSON byte slice.
func EncodeLikeComment(lc LikeComment) ([]byte, error) {
	return json.Marshal(lc)
}

// DecodeLikeComment decodes a JSON byte slice into a LikeComment.
func DecodeLikeComment(payload []byte) (*LikeComment, error) {
	var lc LikeComment

	err := json.Unmarshal(payload, &lc)
	if err != nil {
		return nil, err
	}

	return &lc, nil
}

// LikeCommentReply returns the result of an up or down vote.
type LikeCommentReply struct {
	Total   uint64 `json:"total"`           // Total number of up and down votes
	Result  int64  `json:"result"`          // Current tally of likes, can be negative
	Receipt string `json:"receipt"`         // Server signature of client signature
	Error   string `json:"error,omitempty"` // Error if something wen't wrong during liking a comment
}

// EncodeLikeCommentReply encodes LikeCommentReply into a JSON byte slice.
func EncodeLikeCommentReply(lcr LikeCommentReply) ([]byte, error) {
	return json.Marshal(lcr)
}

// DecodeLikeCommentReply decodes a JSON byte slice into a LikeCommentReply.
func DecodeLikeCommentReply(payload []byte) (*LikeCommentReply, error) {
	var lcr LikeCommentReply

	err := json.Unmarshal(payload, &lcr)
	if err != nil {
		return nil, err
	}

	return &lcr, nil
}

// CensorComment is a journal entry for a censored comment.  The signature and
// public key are from the admin that censored this comment.
type CensorComment struct {
	Token     string `json:"token"`     // Proposal censorship token
	CommentID string `json:"commentid"` // Comment ID
	Reason    string `json:"reason"`    // Reason comment was censored
	Signature string `json:"signature"` // Client signature of Token+CommentID+Reason
	PublicKey string `json:"publickey"` // Pubkey used for signature

	// Generated by decredplugin
	Receipt   string `json:"receipt,omitempty"`   // Server signature of client signature
	Timestamp int64  `json:"timestamp,omitempty"` // Received UNIX timestamp
}

// EncodeCensorComment encodes CensorComment into a JSON byte slice.
func EncodeCensorComment(cc CensorComment) ([]byte, error) {
	return json.Marshal(cc)
}

// DecodeCensorComment decodes a JSON byte slice into a CensorComment.
func DecodeCensorComment(payload []byte) (*CensorComment, error) {
	var cc CensorComment
	err := json.Unmarshal(payload, &cc)
	if err != nil {
		return nil, err
	}
	return &cc, nil
}

// CommentCensorReply returns the receipt for the censoring action. The
// receipt is the server side signature of CommentCensor.Signature.
type CensorCommentReply struct {
	Receipt string `json:"receipt"` // Server signature of client signature
}

// EncodeCensorCommentReply encodes CensorCommentReply into a JSON byte slice.
func EncodeCensorCommentReply(ccr CensorCommentReply) ([]byte, error) {
	return json.Marshal(ccr)
}

// DecodeCensorComment decodes a JSON byte slice into a CensorCommentReply.
func DecodeCensorCommentReply(payload []byte) (*CensorCommentReply, error) {
	var ccr CensorCommentReply
	err := json.Unmarshal(payload, &ccr)
	if err != nil {
		return nil, err
	}
	return &ccr, nil
}

// GetComment retrieves a single comment. The comment can be retrieved by
// either comment ID or by signature.
type GetComment struct {
	Token     string `json:"token"`               // Proposal ID
	CommentID string `json:"commentid,omitempty"` // Comment ID
	Signature string `json:"signature,omitempty"` // Client signature
}

// EncodeGetComment encodes a GetComment into a JSON byte slice.
func EncodeGetComment(gc GetComment) ([]byte, error) {
	return json.Marshal(gc)
}

// DecodeGetComment decodes a JSON byte slice into a GetComment.
func DecodeGetComment(payload []byte) (*GetComment, error) {
	var gc GetComment

	err := json.Unmarshal(payload, &gc)
	if err != nil {
		return nil, err
	}

	return &gc, nil
}

// GetCommentReply returns the provided comment.
type GetCommentReply struct {
	Comment Comment `json:"comment"` // Comment
}

// EncodeGetCommentReply encodes a GetCommentReply into a JSON byte slice.
func EncodeGetCommentReply(gcr GetCommentReply) ([]byte, error) {
	return json.Marshal(gcr)
}

// DecodeGetCommentReply decodes a JSON byte slice into a GetCommentReply.
func DecodeGetCommentReply(payload []byte) (*GetCommentReply, error) {
	var gcr GetCommentReply

	err := json.Unmarshal(payload, &gcr)
	if err != nil {
		return nil, err
	}

	return &gcr, nil
}

// GetComments retrieve all comments for a given proposal. This call returns
// the cooked comments; deleted/censored comments are not returned.
type GetComments struct {
	Token string `json:"token"` // Proposal ID
}

// EncodeGetComments encodes GetCommentsReply into a JSON byte slice.
func EncodeGetComments(gc GetComments) ([]byte, error) {
	return json.Marshal(gc)
}

// DecodeGetComments decodes a JSON byte slice into a GetComments.
func DecodeGetComments(payload []byte) (*GetComments, error) {
	var gc GetComments

	err := json.Unmarshal(payload, &gc)
	if err != nil {
		return nil, err
	}

	return &gc, nil
}

// GetCommentsReply returns the provided number of comments.
type GetCommentsReply struct {
	Comments []Comment `json:"comments"` // Comments
}

// EncodeGetCommentsReply encodes GetCommentsReply into a JSON byte slice.
func EncodeGetCommentsReply(gcr GetCommentsReply) ([]byte, error) {
	return json.Marshal(gcr)
}

// DecodeGetCommentsReply decodes a JSON byte slice into a GetCommentsReply.
func DecodeGetCommentsReply(payload []byte) (*GetCommentsReply, error) {
	var gcr GetCommentsReply

	err := json.Unmarshal(payload, &gcr)
	if err != nil {
		return nil, err
	}

	return &gcr, nil
}

// GetNumComments retrieve the number of comments for a list of proposals.
type GetNumComments struct {
	Tokens []string `json:"tokens"` // Proposal ID
}

// EncodeGetNumComments encodes GetBatchComments into a JSON byte slice.
func EncodeGetNumComments(gnc GetNumComments) ([]byte, error) {
	return json.Marshal(gnc)
}

// DecodeGetNumComments decodes a JSON byte slice into a GetBatchComments.
func DecodeGetNumComments(payload []byte) (*GetNumComments, error) {
	var gnc GetNumComments

	err := json.Unmarshal(payload, &gnc)
	if err != nil {
		return nil, err
	}

	return &gnc, nil
}

// GetNumCommentsReply returns a map from proposal token to int
type GetNumCommentsReply struct {
	CommentsMap map[string]int `json:"commentsmap"`
}

// EncodeGetNumCommentsReply encodes GetNumCommentsReply into a
// JSON byte slice.
func EncodeGetNumCommentsReply(gncr GetNumCommentsReply) ([]byte, error) {
	return json.Marshal(gncr)
}

// DecodeGetNumCommentsReply decodes a JSON byte slice into a
// GetNumCommentsReply.
func DecodeGetNumCommentsReply(payload []byte) (*GetNumCommentsReply, error) {
	var gncr GetNumCommentsReply

	err := json.Unmarshal(payload, &gncr)
	if err != nil {
		return nil, err
	}

	return &gncr, nil
}

// CommentLikes is used to retrieve all of the comment likes for a single
// record comment.
type CommentLikes struct {
	Token     string `json:"token"`     // Censorship token
	CommentID string `json:"commentid"` // Comment ID
}

// EncodeCommentLikes encodes CommentLikes into a JSON byte slice.
func EncodeCommentLikes(gpcv CommentLikes) ([]byte, error) {
	return json.Marshal(gpcv)
}

// DecodeCommentLikes decodes a JSON byte slice into a CommentLikes.
func DecodeCommentLikes(payload []byte) (*CommentLikes, error) {
	var cl CommentLikes

	err := json.Unmarshal(payload, &cl)
	if err != nil {
		return nil, err
	}

	return &cl, nil
}

// CommentLikesReply is the reply to CommentLikes and returns all of the
// upvote/downvote actions for the specified comment.
type CommentLikesReply struct {
	CommentLikes []LikeComment `json:"commentlikes"`
}

// EncodeCommentLikesReply encodes EncodeCommentLikesReply into a JSON byte
// slice.
func EncodeCommentLikesReply(clr CommentLikesReply) ([]byte, error) {
	return json.Marshal(clr)
}

// DecodeCommentLikesReply decodes a JSON byte slice into a CommentLikesReply.
func DecodeCommentLikesReply(payload []byte) (*CommentLikesReply, error) {
	var clr CommentLikesReply

	err := json.Unmarshal(payload, &clr)
	if err != nil {
		return nil, err
	}

	return &clr, nil
}

// GetProposalCommentsLikes is a command to fetch all vote actions
// on the comments of a given proposal
type GetProposalCommentsLikes struct {
	Token string `json:"token"` // Censorship token
}

// EncodeGetProposalCommentsLikes encodes GetProposalCommentsLikes into a JSON byte slice.
func EncodeGetProposalCommentsLikes(gpcv GetProposalCommentsLikes) ([]byte, error) {
	return json.Marshal(gpcv)
}

// DecodeGetProposalCommentsLikes decodes a JSON byte slice into a GetProposalCommentsLikes.
func DecodeGetProposalCommentsLikes(payload []byte) (*GetProposalCommentsLikes, error) {
	var gpcl GetProposalCommentsLikes

	err := json.Unmarshal(payload, &gpcl)
	if err != nil {
		return nil, err
	}

	return &gpcl, nil
}

// GetProposalCommentsLikesReply is a reply with all vote actions
// for the comments of a given proposal
type GetProposalCommentsLikesReply struct {
	CommentsLikes []LikeComment `json:"commentslikes"`
}

// EncodeGetProposalCommentsLikesReply encodes EncodeGetProposalCommentsLikesReply into a JSON byte slice.
func EncodeGetProposalCommentsLikesReply(gpclr GetProposalCommentsLikesReply) ([]byte, error) {
	return json.Marshal(gpclr)
}

// DecodeGetProposalCommentsLikesReply decodes a JSON byte slice into a GetProposalCommentsLikesReply.
func DecodeGetProposalCommentsLikesReply(payload []byte) (*GetProposalCommentsLikesReply, error) {
	var gpclr GetProposalCommentsLikesReply

	err := json.Unmarshal(payload, &gpclr)
	if err != nil {
		return nil, err
	}

	return &gpclr, nil
}

// Inventory is used to retrieve the decred plugin inventory.
type Inventory struct{}

// EncodeInventory encodes Inventory into a JSON byte slice.
func EncodeInventory(i Inventory) ([]byte, error) {
	return json.Marshal(i)
}

// DecodeInventory decodes a JSON byte slice into a Inventory.
func DecodeInventory(payload []byte) (*Inventory, error) {
	var i Inventory

	err := json.Unmarshal(payload, &i)
	if err != nil {
		return nil, err
	}

	return &i, nil
}

// StartVoteTuple is used to return the StartVote and StartVoteReply for a
// record. StartVoteReply does not contain any record identifying data so it
// must be returned with the StartVote in order to know what record it belongs
// to.
type StartVoteTuple struct {
	StartVote      StartVote      `json:"startvote"`      // Start vote
	StartVoteReply StartVoteReply `json:"startvotereply"` // Start vote reply
}

// InventoryReply returns the decred plugin inventory.
type InventoryReply struct {
	Comments             []Comment            `json:"comments"`             // Comments
	LikeComments         []LikeComment        `json:"likecomments"`         // Like comments
	AuthorizeVotes       []AuthorizeVote      `json:"authorizevotes"`       // Authorize votes
	AuthorizeVoteReplies []AuthorizeVoteReply `json:"authorizevotereplies"` // Authorize vote replies
	StartVoteTuples      []StartVoteTuple     `json:"startvotetuples"`      // Start vote tuples
	CastVotes            []CastVote           `json:"castvotes"`            // Cast votes
}

// EncodeInventoryReply encodes a InventoryReply into a JSON byte slice.
func EncodeInventoryReply(ir InventoryReply) ([]byte, error) {
	return json.Marshal(ir)
}

// DecodeInventoryReply decodes a JSON byte slice into a inventory.
func DecodeInventoryReply(payload []byte) (*InventoryReply, error) {
	var ir InventoryReply

	err := json.Unmarshal(payload, &ir)
	if err != nil {
		return nil, err
	}

	return &ir, nil
}

// TokenInventory requests the tokens of the records in the inventory,
// categorized by stage of the voting process. By default, only vetted
// records are returned.
type TokenInventory struct {
	BestBlock uint64 `json:"bestblock"` // Best block
	Unvetted  bool   `json:"unvetted"`  // Include unvetted records
}

// EncodeTokenInventory encodes a TokenInventory into a JSON byte slice.
func EncodeTokenInventory(i TokenInventory) ([]byte, error) {
	return json.Marshal(i)
}

// DecodeTokenInventory decodes a JSON byte slice into a TokenInventory.
func DecodeTokenInventory(payload []byte) (*TokenInventory, error) {
	var i TokenInventory

	err := json.Unmarshal(payload, &i)
	if err != nil {
		return nil, err
	}

	return &i, nil
}

// TokenInventoryReply is the response to the TokenInventory command and
// returns the tokens of all records in the inventory. The tokens are
// categorized by stage of the voting process and are sorted according to
// the following rule.
//
// Sorted by record timestamp in descending order:
// Pre, Abandonded, Unreviewed, Censored
//
// Sorted by voting period end block height in descending order:
// Active, Approved, Rejected
type TokenInventoryReply struct {
	// Vetted Records
	Pre       []string `json:"pre"`       // Tokens of records that are pre-vote
	Active    []string `json:"active"`    // Tokens of records with an active voting period
	Approved  []string `json:"approved"`  // Tokens of records that have been approved by a vote
	Rejected  []string `json:"rejected"`  // Tokens of records that have been rejected by a vote
	Abandoned []string `json:"abandoned"` // Tokens of records that have been abandoned

	// Unvetted records
	Unreviewed []string `json:"unreviewied,omitempty"` // Tokens of records that are unreviewed
	Censored   []string `json:"censored,omitempty"`    // Tokens of records that have been censored
}

// EncodeTokenInventoryReply encodes a TokenInventoryReply into a JSON byte
// slice.
func EncodeTokenInventoryReply(itr TokenInventoryReply) ([]byte, error) {
	return json.Marshal(itr)
}

// DecodeTokenInventoryReply decodes a JSON byte slice into a inventory.
func DecodeTokenInventoryReply(payload []byte) (*TokenInventoryReply, error) {
	var itr TokenInventoryReply

	err := json.Unmarshal(payload, &itr)
	if err != nil {
		return nil, err
	}

	return &itr, nil
}

// LoadVoteResults creates a vote results entry in the cache for any proposals
// that have finsished voting but have not yet been added to the lazy loaded
// vote results table.
type LoadVoteResults struct {
	BestBlock uint64 `json:"bestblock"` // Best block height
}

// EncodeLoadVoteResults encodes a LoadVoteResults into a JSON byte slice.
func EncodeLoadVoteResults(lvr LoadVoteResults) ([]byte, error) {
	return json.Marshal(lvr)
}

// DecodeLoadVoteResults decodes a JSON byte slice into a LoadVoteResults.
func DecodeLoadVoteResults(payload []byte) (*LoadVoteResults, error) {
	var lvr LoadVoteResults

	err := json.Unmarshal(payload, &lvr)
	if err != nil {
		return nil, err
	}

	return &lvr, nil
}

// LoadVoteResultsReply is the reply to the LoadVoteResults command.
type LoadVoteResultsReply struct{}

// EncodeLoadVoteResultsReply encodes a LoadVoteResultsReply into a JSON
// byte slice.
func EncodeLoadVoteResultsReply(reply LoadVoteResultsReply) ([]byte, error) {
	return json.Marshal(reply)
}

// DecodeLoadVoteResultsReply decodes a JSON byte slice into a LoadVoteResults.
func DecodeLoadVoteResultsReply(payload []byte) (*LoadVoteResultsReply, error) {
	var reply LoadVoteResultsReply

	err := json.Unmarshal(payload, &reply)
	if err != nil {
		return nil, err
	}

	return &reply, nil
}
